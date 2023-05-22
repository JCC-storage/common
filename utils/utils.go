package utils

import "fmt"

// MakeMoveOperationFileName Move操作时，写入的文件的名称
func MakeMoveOperationFileName(objectID int, userID int) string {
	return fmt.Sprintf("%d-%d", objectID, userID)
}
