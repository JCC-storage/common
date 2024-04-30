package prescheduler

import (
	"fmt"
	"sort"

	"github.com/samber/lo"

	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	uopsdk "gitlink.org.cn/cloudream/common/sdks/unifyops"
	schglb "gitlink.org.cn/cloudream/scheduler/common/globals"
	schmod "gitlink.org.cn/cloudream/scheduler/common/models"
	jobmod "gitlink.org.cn/cloudream/scheduler/common/models/job"
	mgrmq "gitlink.org.cn/cloudream/scheduler/common/pkgs/mq/manager"
)

const (
	//每个节点划分的资源等级：
	// ResourceLevel1：表示所有资源类型均满足 大于等于1.5倍
	ResourceLevel1 = 1
	// ResourceLevel2：表示不满足Level1，但所有资源类型均满足 大于等于1倍
	ResourceLevel2 = 2
	// ResourceLevel3： 表示某些资源类型 小于一倍
	ResourceLevel3 = 3

	CpuResourceWeight float64 = 1
	StgResourceWeight float64 = 1.2

	CachingWeight float64 = 1
	LoadedWeight  float64 = 2
)

var ErrNoAvailableScheme = fmt.Errorf("no appropriate scheduling node found, please wait")

type candidate struct {
	CC                    schmod.ComputingCenter
	IsReferencedJobTarget bool // 这个节点是否是所依赖的任务所选择的节点
	Resource              resourcesDetail
	Files                 filesDetail
}

type resourcesDetail struct {
	CPU     resourceDetail
	GPU     resourceDetail
	NPU     resourceDetail
	MLU     resourceDetail
	Storage resourceDetail
	Memory  resourceDetail

	TotalScore float64
	AvgScore   float64
	MaxLevel   int
}
type resourceDetail struct {
	Level int
	Score float64
}

type filesDetail struct {
	Dataset fileDetail
	Code    fileDetail
	Image   fileDetail

	TotalScore float64
}
type fileDetail struct {
	CachingScore float64
	LoadingScore float64
	IsLoaded     bool //表示storage是否已经调度到该节点, image表示镜像是否已经加载到该算力中心
}

type schedulingJob struct {
	Job    schsdk.JobInfo
	Afters []string
}

type CandidateArr []*candidate

func (a CandidateArr) Len() int      { return len(a) }
func (a CandidateArr) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a CandidateArr) Less(i, j int) bool {
	n1 := a[i]
	n2 := a[j]

	// 优先与所依赖的任务放到一起，但要求那个节点的资源足够
	if n1.IsReferencedJobTarget && n1.Resource.MaxLevel < ResourceLevel3 {
		return true
	}
	if n2.IsReferencedJobTarget && n2.Resource.MaxLevel < ResourceLevel3 {
		return true
	}

	// 优先判断资源等级，资源等级越低，代表越满足需求
	if n1.Resource.MaxLevel < n2.Resource.MaxLevel {
		return true
	}
	if n1.Resource.MaxLevel > n2.Resource.MaxLevel {
		return false
	}

	// 等级相同时，根据单项分值比较
	switch n1.Resource.MaxLevel {
	case ResourceLevel1:
		// 数据文件总分越高，代表此节点上拥有的数据文件越完整，则越优先考虑
		return n1.Files.TotalScore > n2.Files.TotalScore

	case ResourceLevel2:
		// 资源分的平均值越高，代表资源越空余，则越优先考虑
		return n1.Resource.AvgScore > n2.Resource.AvgScore

	case ResourceLevel3:
		// 资源分的平均值越高，代表资源越空余，则越优先考虑
		return n1.Resource.AvgScore > n2.Resource.AvgScore
	}

	return false
}

type DefaultPreScheduler struct {
}

func NewDefaultPreScheduler() *DefaultPreScheduler {
	return &DefaultPreScheduler{}
}

// ScheduleJobSet 任务集预调度
func (s *DefaultPreScheduler) ScheduleJobSet(info *schsdk.JobSetInfo) (*jobmod.JobSetPreScheduleScheme, *schsdk.JobSetFilesUploadScheme, error) {
	jobSetScheme := &jobmod.JobSetPreScheduleScheme{
		JobSchemes: make(map[string]jobmod.JobScheduleScheme),
	}
	filesUploadSchemes := make(map[string]schsdk.LocalFileUploadScheme)

	mgrCli, err := schglb.ManagerMQPool.Acquire()
	if err != nil {
		return nil, nil, fmt.Errorf("new collector client: %w", err)
	}
	defer schglb.ManagerMQPool.Release(mgrCli)

	// 查询有哪些算力中心可用

	allCC, err := mgrCli.GetAllComputingCenter(mgrmq.NewGetAllComputingCenter())
	if err != nil {
		return nil, nil, fmt.Errorf("getting all computing center info: %w", err)
	}

	ccs := make(map[schsdk.CCID]schmod.ComputingCenter)
	for _, node := range allCC.ComputingCenters {
		ccs[node.CCID] = node
	}

	if len(ccs) == 0 {
		return nil, nil, ErrNoAvailableScheme
	}

	// 先根据任务配置，收集它们依赖的任务的LocalID
	var schJobs []*schedulingJob
	for _, job := range info.Jobs {
		j := &schedulingJob{
			Job: job,
		}

		if norJob, ok := job.(*schsdk.NormalJobInfo); ok {
			if resFile, ok := norJob.Files.Dataset.(*schsdk.DataReturnJobFileInfo); ok {
				j.Afters = append(j.Afters, resFile.DataReturnLocalJobID)
			}

			if resFile, ok := norJob.Files.Code.(*schsdk.DataReturnJobFileInfo); ok {
				j.Afters = append(j.Afters, resFile.DataReturnLocalJobID)
			}
		} else if resJob, ok := job.(*schsdk.DataReturnJobInfo); ok {
			j.Afters = append(j.Afters, resJob.TargetLocalJobID)
		}

		schJobs = append(schJobs, j)
	}

	// 然后根据依赖进行排序
	schJobs, ok := s.orderByAfters(schJobs)
	if !ok {
		return nil, nil, fmt.Errorf("circular reference detected between jobs in the job set")
	}

	// 经过排序后，按顺序生成调度方案
	for _, job := range schJobs {
		if norJob, ok := job.Job.(*schsdk.NormalJobInfo); ok {
			scheme, err := s.scheduleForNormalJob(info, job, ccs, jobSetScheme.JobSchemes)
			if err != nil {
				return nil, nil, err
			}

			jobSetScheme.JobSchemes[job.Job.GetLocalJobID()] = *scheme

			// 检查数据文件的配置项，生成上传文件方案
			s.fillNormarlJobLocalUploadScheme(norJob, scheme.TargetCCID, filesUploadSchemes, ccs)
		}

		// 回源任务目前不需要生成调度方案
	}

	return jobSetScheme, &schsdk.JobSetFilesUploadScheme{
		LocalFileSchemes: lo.Values(filesUploadSchemes),
	}, nil
}

// ScheduleJob 单个任务预调度
func (s *DefaultPreScheduler) ScheduleJob(instJobInfo *schsdk.InstanceJobInfo) (*jobmod.JobScheduleScheme, *schsdk.JobFilesUploadScheme, error) {
	filesUploadSchemes := make(map[string]schsdk.LocalFileUploadScheme)

	mgrCli, err := schglb.ManagerMQPool.Acquire()
	if err != nil {
		return nil, nil, fmt.Errorf("new collector client: %w", err)
	}
	defer schglb.ManagerMQPool.Release(mgrCli)

	// 查询有哪些算力中心可用

	allCC, err := mgrCli.GetAllComputingCenter(mgrmq.NewGetAllComputingCenter())
	if err != nil {
		return nil, nil, fmt.Errorf("getting all computing center info: %w", err)
	}

	ccs := make(map[schsdk.CCID]schmod.ComputingCenter)
	for _, node := range allCC.ComputingCenters {
		ccs[node.CCID] = node
	}

	if len(ccs) == 0 {
		return nil, nil, ErrNoAvailableScheme
	}

	info := &schsdk.NormalJobInfo{
		Files:     instJobInfo.Files,
		Runtime:   instJobInfo.Runtime,
		Resources: instJobInfo.Resources,
	}

	job := &schedulingJob{
		Job: info,
	}
	scheme, err := s.scheduleForNormalJob2(job, ccs)
	if err != nil {
		return nil, nil, err
	}

	// 检查数据文件的配置项，生成上传文件方案
	s.fillNormarlJobLocalUploadScheme(info, scheme.TargetCCID, filesUploadSchemes, ccs)

	return scheme, &schsdk.JobFilesUploadScheme{
		LocalFileSchemes: lo.Values(filesUploadSchemes),
	}, nil
}

func (s *DefaultPreScheduler) orderByAfters(jobs []*schedulingJob) ([]*schedulingJob, bool) {
	type jobOrder struct {
		Job    *schedulingJob
		Afters []string
	}

	var jobOrders []*jobOrder
	for _, job := range jobs {
		od := &jobOrder{
			Job:    job,
			Afters: make([]string, len(job.Afters)),
		}

		copy(od.Afters, job.Afters)

		jobOrders = append(jobOrders, od)
	}

	// 然后排序
	var orderedJob []*schedulingJob
	for {
		rm := 0
		for i, jo := range jobOrders {
			// 找到没有依赖的任务，然后将其取出
			if len(jo.Afters) == 0 {
				orderedJob = append(orderedJob, jo.Job)

				// 删除其他任务对它的引用
				for _, job2 := range jobOrders {
					job2.Afters = lo.Reject(job2.Afters, func(item string, idx int) bool { return item == jo.Job.Job.GetLocalJobID() })
				}

				rm++
				continue
			}

			jobOrders[i-rm] = jobOrders[i]
		}

		jobOrders = jobOrders[:len(jobOrders)-rm]
		if len(jobOrders) == 0 {
			break
		}

		// 遍历一轮后没有找到无依赖的任务，那么就是存在循环引用，排序失败
		if rm == 0 {
			return nil, false
		}
	}

	return orderedJob, true
}

func (s *DefaultPreScheduler) scheduleForNormalJob(jobSet *schsdk.JobSetInfo, job *schedulingJob, ccs map[schsdk.CCID]schmod.ComputingCenter, jobSchemes map[string]jobmod.JobScheduleScheme) (*jobmod.JobScheduleScheme, error) {
	allCCs := make(map[schsdk.CCID]*candidate)

	// 初始化备选节点信息
	for _, cc := range ccs {
		caNode := &candidate{
			CC: cc,
		}

		// 检查此节点是否是它所引用的任务所选的节点
		for _, af := range job.Afters {
			resJob := findJobInfo[*schsdk.DataReturnJobInfo](jobSet.Jobs, af)
			if resJob == nil {
				return nil, fmt.Errorf("resource job %s not found in the job set", af)
			}

			// 由于jobs已经按照引用排序，所以正常情况下这里肯定能取到值
			scheme, ok := jobSchemes[resJob.TargetLocalJobID]
			if !ok {
				continue
			}

			if scheme.TargetCCID == cc.CCID {
				caNode.IsReferencedJobTarget = true
				break
			}
		}

		allCCs[cc.CCID] = caNode
	}

	norJob := job.Job.(*schsdk.NormalJobInfo)

	// 计算文件占有量得分
	err := s.calcFileScore(norJob.Files, allCCs)
	if err != nil {
		return nil, err
	}

	// 计算资源余量得分
	err = s.calcResourceScore(norJob, allCCs)
	if err != nil {
		return nil, err
	}

	allCCsArr := lo.Values(allCCs)
	sort.Sort(CandidateArr(allCCsArr))

	targetNode := allCCsArr[0]
	if targetNode.Resource.MaxLevel == ResourceLevel3 {
		return nil, ErrNoAvailableScheme
	}

	scheme := s.makeSchemeForNode(norJob, targetNode)
	return &scheme, nil
}

func (s *DefaultPreScheduler) scheduleForNormalJob2(job *schedulingJob, ccs map[schsdk.CCID]schmod.ComputingCenter) (*jobmod.JobScheduleScheme, error) {
	allCCs := make(map[schsdk.CCID]*candidate)

	// 初始化备选节点信息
	for _, cc := range ccs {
		caNode := &candidate{
			CC: cc,
		}

		allCCs[cc.CCID] = caNode
	}

	norJob := job.Job.(*schsdk.NormalJobInfo)

	// 计算文件占有量得分
	err := s.calcFileScore(norJob.Files, allCCs)
	if err != nil {
		return nil, err
	}

	// 计算资源余量得分
	err = s.calcResourceScore(norJob, allCCs)
	if err != nil {
		return nil, err
	}

	allCCsArr := lo.Values(allCCs)
	sort.Sort(CandidateArr(allCCsArr))

	targetNode := allCCsArr[0]
	if targetNode.Resource.MaxLevel == ResourceLevel3 {
		return nil, ErrNoAvailableScheme
	}

	scheme := s.makeSchemeForNode(norJob, targetNode)
	return &scheme, nil
}

func (s *DefaultPreScheduler) fillNormarlJobLocalUploadScheme(norJob *schsdk.NormalJobInfo, targetCCID schsdk.CCID, schemes map[string]schsdk.LocalFileUploadScheme, ccs map[schsdk.CCID]schmod.ComputingCenter) {
	if localFile, ok := norJob.Files.Dataset.(*schsdk.LocalJobFileInfo); ok {
		if _, ok := schemes[localFile.LocalPath]; !ok {
			cdsNodeID := ccs[targetCCID].CDSNodeID
			schemes[localFile.LocalPath] = schsdk.LocalFileUploadScheme{
				LocalPath:         localFile.LocalPath,
				UploadToCDSNodeID: &cdsNodeID,
			}
		}
	}

	if localFile, ok := norJob.Files.Code.(*schsdk.LocalJobFileInfo); ok {
		if _, ok := schemes[localFile.LocalPath]; !ok {
			cdsNodeID := ccs[targetCCID].CDSNodeID
			schemes[localFile.LocalPath] = schsdk.LocalFileUploadScheme{
				LocalPath:         localFile.LocalPath,
				UploadToCDSNodeID: &cdsNodeID,
			}
		}
	}

	if localFile, ok := norJob.Files.Image.(*schsdk.LocalJobFileInfo); ok {
		if _, ok := schemes[localFile.LocalPath]; !ok {
			cdsNodeID := ccs[targetCCID].CDSNodeID
			schemes[localFile.LocalPath] = schsdk.LocalFileUploadScheme{
				LocalPath:         localFile.LocalPath,
				UploadToCDSNodeID: &cdsNodeID,
			}
		}
	}
}

func (s *DefaultPreScheduler) makeSchemeForNode(job *schsdk.NormalJobInfo, targetCC *candidate) jobmod.JobScheduleScheme {
	scheme := jobmod.JobScheduleScheme{
		TargetCCID: targetCC.CC.CCID,
	}

	// TODO 根据实际情况选择Move或者Load

	if _, ok := job.Files.Dataset.(*schsdk.PackageJobFileInfo); ok && !targetCC.Files.Dataset.IsLoaded {
		scheme.Dataset.Action = jobmod.ActionLoad
	} else {
		scheme.Dataset.Action = jobmod.ActionNo
	}

	if _, ok := job.Files.Code.(*schsdk.PackageJobFileInfo); ok && !targetCC.Files.Code.IsLoaded {
		scheme.Code.Action = jobmod.ActionLoad
	} else {
		scheme.Code.Action = jobmod.ActionNo
	}

	if _, ok := job.Files.Image.(*schsdk.PackageJobFileInfo); ok && !targetCC.Files.Image.IsLoaded {
		scheme.Image.Action = jobmod.ActionImportImage
	} else {
		scheme.Image.Action = jobmod.ActionNo
	}

	return scheme
}

func findResuorce[T uopsdk.ResourceData](all []uopsdk.ResourceData) T {
	for _, data := range all {
		if ret, ok := data.(T); ok {
			return ret
		}
	}

	var def T
	return def
}

func findJobInfo[T schsdk.JobInfo](jobs []schsdk.JobInfo, localJobID string) T {
	for _, job := range jobs {
		if ret, ok := job.(T); ok && job.GetLocalJobID() == localJobID {
			return ret
		}
	}

	var def T
	return def
}
