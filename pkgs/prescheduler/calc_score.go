package prescheduler

import (
	"fmt"
	"github.com/inhies/go-bytesize"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	uopsdk "gitlink.org.cn/cloudream/common/sdks/unifyops"
	"gitlink.org.cn/cloudream/common/utils/math2"
	schglb "gitlink.org.cn/cloudream/scheduler/common/globals"
	schmod "gitlink.org.cn/cloudream/scheduler/common/models"
	"gitlink.org.cn/cloudream/scheduler/common/pkgs/mq/collector"
	mgrmq "gitlink.org.cn/cloudream/scheduler/common/pkgs/mq/manager"
)

func (s *DefaultPreScheduler) calcResourceScore(job *schsdk.NormalJobInfo, allCCs map[schsdk.CCID]*candidate) error {
	for _, cc := range allCCs {
		res, err := s.calcOneResourceScore(job.Resources, &cc.CC)
		if err != nil {
			return err
		}

		cc.Resource = *res
	}

	return nil
}

// 划分节点资源等级，并计算资源得分
func (s *DefaultPreScheduler) calcOneResourceScore(requires schsdk.JobResourcesInfo, cc *schmod.ComputingCenter) (*resourcesDetail, error) {
	colCli, err := schglb.CollectorMQPool.Acquire()
	if err != nil {
		return nil, fmt.Errorf("new collector client: %w", err)
	}
	defer schglb.CollectorMQPool.Release(colCli)

	getResDataResp, err := colCli.GetAllResourceData(collector.NewGetAllResourceData(cc.UOPSlwNodeID))
	if err != nil {
		return nil, err
	}

	var resDetail resourcesDetail

	//计算资源得分
	totalScore := 0.0
	maxLevel := 0
	resKinds := 0

	if requires.CPU > 0 {
		res := findResuorce[*uopsdk.CPUResourceData](getResDataResp.Datas)
		if res == nil {
			resDetail.CPU.Level = ResourceLevel3
			resDetail.CPU.Score = 0
		} else {
			resDetail.CPU.Level = s.calcResourceLevel(float64(res.Available.Value), requires.CPU)
			resDetail.CPU.Score = (float64(res.Available.Value) / requires.CPU) * CpuResourceWeight
		}

		maxLevel = math2.Max(maxLevel, resDetail.CPU.Level)
		totalScore += resDetail.CPU.Score
		resKinds++
	}

	if requires.GPU > 0 {
		res := findResuorce[*uopsdk.GPUResourceData](getResDataResp.Datas)
		if res == nil {
			resDetail.GPU.Level = ResourceLevel3
			resDetail.GPU.Score = 0
		} else {
			resDetail.GPU.Level = s.calcResourceLevel(float64(res.Available.Value), requires.GPU)
			resDetail.GPU.Score = (float64(res.Available.Value) / requires.GPU) * CpuResourceWeight
		}

		maxLevel = math2.Max(maxLevel, resDetail.GPU.Level)
		totalScore += resDetail.GPU.Score
		resKinds++
	}

	if requires.NPU > 0 {
		res := findResuorce[*uopsdk.NPUResourceData](getResDataResp.Datas)
		if res == nil {
			resDetail.NPU.Level = ResourceLevel3
			resDetail.NPU.Score = 0
		} else {
			resDetail.NPU.Level = s.calcResourceLevel(float64(res.Available.Value), requires.NPU)
			resDetail.NPU.Score = (float64(res.Available.Value) / requires.NPU) * CpuResourceWeight
		}

		maxLevel = math2.Max(maxLevel, resDetail.NPU.Level)
		totalScore += resDetail.NPU.Score
		resKinds++
	}

	if requires.MLU > 0 {
		res := findResuorce[*uopsdk.MLUResourceData](getResDataResp.Datas)
		if res == nil {
			resDetail.MLU.Level = ResourceLevel3
			resDetail.MLU.Score = 0
		} else {
			resDetail.MLU.Level = s.calcResourceLevel(float64(res.Available.Value), requires.MLU)
			resDetail.MLU.Score = (float64(res.Available.Value) / requires.MLU) * CpuResourceWeight
		}

		maxLevel = math2.Max(maxLevel, resDetail.MLU.Level)
		totalScore += resDetail.MLU.Score
		resKinds++
	}

	if requires.Storage > 0 {
		res := findResuorce[*uopsdk.StorageResourceData](getResDataResp.Datas)
		if res == nil {
			resDetail.Storage.Level = ResourceLevel3
			resDetail.Storage.Score = 0
		} else {
			bytes, err := bytesize.Parse(fmt.Sprintf("%f%s", res.Available.Value, res.Available.Unit))
			if err != nil {
				return nil, err
			}

			resDetail.Storage.Level = s.calcResourceLevel(float64(bytes), float64(requires.Storage))
			resDetail.Storage.Score = (float64(bytes) / float64(requires.Storage)) * StgResourceWeight
		}

		maxLevel = math2.Max(maxLevel, resDetail.Storage.Level)
		totalScore += resDetail.Storage.Score
		resKinds++
	}

	if requires.Memory > 0 {
		res := findResuorce[*uopsdk.MemoryResourceData](getResDataResp.Datas)
		if res == nil {
			resDetail.Memory.Level = ResourceLevel3
			resDetail.Memory.Score = 0
		} else {
			bytes, err := bytesize.Parse(fmt.Sprintf("%f%s", res.Available.Value, res.Available.Unit))
			if err != nil {
				return nil, err
			}

			resDetail.Memory.Level = s.calcResourceLevel(float64(bytes), float64(requires.Memory))
			resDetail.Memory.Score = (float64(bytes) / float64(requires.Memory)) * StgResourceWeight
		}

		maxLevel = math2.Max(maxLevel, resDetail.Memory.Level)
		totalScore += resDetail.Memory.Score
		resKinds++
	}

	if resKinds == 0 {
		return &resDetail, nil
	}

	resDetail.TotalScore = totalScore
	resDetail.AvgScore = resDetail.AvgScore / float64(resKinds)
	resDetail.MaxLevel = maxLevel

	return &resDetail, nil
}

func (s *DefaultPreScheduler) calcResourceLevel(avai float64, need float64) int {
	if avai >= 1.5*need {
		return ResourceLevel1
	}

	if avai >= need {
		return ResourceLevel2
	}

	return ResourceLevel3
}

// 计算节点得分情况
func (s *DefaultPreScheduler) calcFileScore(files schsdk.JobFilesInfo, allCCs map[schsdk.CCID]*candidate) error {
	// 只计算运控返回的可用计算中心上的存储服务的数据权重
	cdsNodeToCC := make(map[cdssdk.NodeID]*candidate)
	for _, cc := range allCCs {
		cdsNodeToCC[cc.CC.CDSNodeID] = cc
	}

	//计算code相关得分
	if pkgFile, ok := files.Code.(*schsdk.PackageJobFileInfo); ok {
		codeFileScores, err := s.calcPackageFileScore(pkgFile.PackageID, cdsNodeToCC)
		if err != nil {
			return fmt.Errorf("calc code file score: %w", err)
		}
		for id, score := range codeFileScores {
			allCCs[id].Files.Code = *score
		}
	}

	//计算dataset相关得分
	if pkgFile, ok := files.Dataset.(*schsdk.PackageJobFileInfo); ok {
		datasetFileScores, err := s.calcPackageFileScore(pkgFile.PackageID, cdsNodeToCC)
		if err != nil {
			return fmt.Errorf("calc dataset file score: %w", err)
		}
		for id, score := range datasetFileScores {
			allCCs[id].Files.Dataset = *score
		}
	}

	//计算image相关得分
	if imgFile, ok := files.Image.(*schsdk.ImageJobFileInfo); ok {
		//计算image相关得分
		imageFileScores, err := s.calcImageFileScore(imgFile.ImageID, allCCs, cdsNodeToCC)
		if err != nil {
			return fmt.Errorf("calc image file score: %w", err)
		}
		for id, score := range imageFileScores {
			allCCs[id].Files.Image = *score
		}
	}

	for _, cc := range allCCs {
		cc.Files.TotalScore = cc.Files.Code.CachingScore +
			cc.Files.Code.LoadingScore +
			cc.Files.Dataset.CachingScore +
			cc.Files.Dataset.LoadingScore +
			cc.Files.Image.CachingScore +
			cc.Files.Image.LoadingScore
	}

	return nil
}

// 计算package在各节点的得分情况
func (s *DefaultPreScheduler) calcPackageFileScore(packageID cdssdk.PackageID, cdsNodeToCC map[cdssdk.NodeID]*candidate) (map[schsdk.CCID]*fileDetail, error) {
	colCli, err := schglb.CollectorMQPool.Acquire()
	if err != nil {
		return nil, fmt.Errorf("new collector client: %w", err)
	}
	defer schglb.CollectorMQPool.Release(colCli)

	ccFileScores := make(map[schsdk.CCID]*fileDetail)

	// TODO UserID
	cachedResp, err := colCli.PackageGetCachedStgNodes(collector.NewPackageGetCachedStgNodes(1, packageID))
	if err != nil {
		return nil, err
	}

	for _, cdsNodeCacheInfo := range cachedResp.NodeInfos {
		cc, ok := cdsNodeToCC[cdsNodeCacheInfo.NodeID]
		if !ok {
			continue
		}

		ccFileScores[cc.CC.CCID] = &fileDetail{
			//TODO 根据缓存方式不同，可能会有不同的计算方式
			CachingScore: float64(cdsNodeCacheInfo.FileSize) / float64(cachedResp.PackageSize) * CachingWeight,
		}
	}

	// TODO UserID
	loadedResp, err := colCli.PackageGetLoadedStgNodes(collector.NewPackageGetLoadedStgNodes(1, packageID))
	if err != nil {
		return nil, err
	}

	for _, cdsNodeID := range loadedResp.StgNodeIDs {
		cc, ok := cdsNodeToCC[cdsNodeID]
		if !ok {
			continue
		}

		sfc, ok := ccFileScores[cc.CC.CCID]
		if !ok {
			sfc = &fileDetail{}
			ccFileScores[cc.CC.CCID] = sfc
		}

		sfc.LoadingScore = 1 * LoadedWeight
		sfc.IsLoaded = true
	}

	return ccFileScores, nil
}

// 计算package在各节点的得分情况
func (s *DefaultPreScheduler) calcImageFileScore(imageID schsdk.ImageID, allCCs map[schsdk.CCID]*candidate, cdsNodeToCC map[cdssdk.NodeID]*candidate) (map[schsdk.CCID]*fileDetail, error) {
	colCli, err := schglb.CollectorMQPool.Acquire()
	if err != nil {
		return nil, fmt.Errorf("new collector client: %w", err)
	}
	defer schglb.CollectorMQPool.Release(colCli)

	magCli, err := schglb.ManagerMQPool.Acquire()
	if err != nil {
		return nil, fmt.Errorf("new manager client: %w", err)
	}
	defer schglb.ManagerMQPool.Release(magCli)

	imageInfoResp, err := magCli.GetImageInfo(mgrmq.NewGetImageInfo(imageID))
	if err != nil {
		return nil, fmt.Errorf("getting image info: %w", err)
	}

	ccFileScores := make(map[schsdk.CCID]*fileDetail)

	if imageInfoResp.Image.CDSPackageID != nil {
		cachedResp, err := colCli.PackageGetCachedStgNodes(collector.NewPackageGetCachedStgNodes(1, *imageInfoResp.Image.CDSPackageID))
		if err != nil {
			return nil, err
		}

		for _, cdsNodeCacheInfo := range cachedResp.NodeInfos {
			cc, ok := cdsNodeToCC[cdsNodeCacheInfo.NodeID]
			if !ok {
				continue
			}

			ccFileScores[cc.CC.CCID] = &fileDetail{
				//TODO 根据缓存方式不同，可能会有不同的计算方式
				CachingScore: float64(cdsNodeCacheInfo.FileSize) / float64(cachedResp.PackageSize) * CachingWeight,
			}
		}
	}

	// 镜像的LoadingScore是判断是否导入到算力中心
	for _, pcmImg := range imageInfoResp.PCMImages {
		_, ok := allCCs[pcmImg.CCID]
		if !ok {
			continue
		}

		fsc, ok := ccFileScores[pcmImg.CCID]
		if !ok {
			fsc = &fileDetail{}
			ccFileScores[pcmImg.CCID] = fsc
		}

		fsc.LoadingScore = 1 * LoadedWeight
		fsc.IsLoaded = true
	}

	return ccFileScores, nil
}
