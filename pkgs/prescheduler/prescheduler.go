package prescheduler

import (
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	jobmod "gitlink.org.cn/cloudream/scheduler/common/models/job"
)

type PreScheduler interface {
	ScheduleJobSet(info *schsdk.JobSetInfo) (*jobmod.JobSetPreScheduleScheme, *schsdk.JobSetFilesUploadScheme, error)
	ScheduleJob(info *schsdk.InstanceJobInfo) (*jobmod.JobScheduleScheme, *schsdk.JobFilesUploadScheme, error)
}
