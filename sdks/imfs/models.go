package imsdk

import schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"

const (
	EnvPackageList    = "IMFS_PACKAGE_LIST"
	EnvServiceAddress = "IMFS_SERVICE_ADDRESS"
	EnvLocalJobID     = "IMFS_LOCAL_JOB_ID"
	EnvJobsetID       = "IMFS_JOBSET_ID"
	EnvListeningList  = "IMFS_PROXY_LSTENING_LIST"
	EnvServingList    = "IMFS_PROXY_SERVING_LIST"
)

//代表本任务需要访问的服务
type ClientService struct {
	Name string `json:"name"`
}

//代表任务给提供各服务的端口
type ServerService struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

type FullJobID struct {
	JobSetID   schsdk.JobSetID
	LocalJobID string
}
