package storage

import "path/filepath"

func MakeIPFSFilePath(fileHash string) string {
	return filepath.Join("ipfs", fileHash)
}
