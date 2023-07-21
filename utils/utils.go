package utils

import "fmt"

// MakeMoveOperationFileName Move操作时，写入的文件的名称
func MakeMoveOperationFileName(objectID int64, userID int64) string {
	return fmt.Sprintf("%d-%d", objectID, userID)
}
