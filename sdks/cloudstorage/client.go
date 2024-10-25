package cloudstorage

import "fmt"

//type ObjectStorageInfo interface {
//	NewClient() (ObjectStorageClient, error)
//}

type ObjectStorageClient interface {
	InitiateMultipartUpload(objectName string) (string, error)
	UploadPart()
	CompleteMultipartUpload() (string, error)
	AbortMultipartUpload()
	Close()
}

func NewObjectStorageClient(info ObjectStorage) (ObjectStorageClient, error) {
	switch info.Manufacturer {
	case AliCloud:
		return NewOSSClient(info), nil
	case HuaweiCloud:
		return &OBSClient{}, nil
	}
	return nil, fmt.Errorf("unknown cloud storage manufacturer %s", info.Manufacturer)
}
