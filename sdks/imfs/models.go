package imsdk

import schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"

const (
	EnvPackageList    = "IMFS_PACKAGE_LIST"
	EnvServiceAddress = "IMFS_SERVICE_ADDRESS"

	EnvLocalJobID        = "LOCAL_JOB_ID"
	EnvJobsetID          = "JOBSET_ID"
	EnvClientServiceList = "CLIENT_SERVICE_LIST"
	EnvServerServiceList = "SERVER_SERVICE_LIST"
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
