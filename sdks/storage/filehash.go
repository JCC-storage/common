package cdssdk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// 文件的哈希值，格式：[前缀: 4个字符][哈希值: 64个字符]
// 前缀用于区分哈希值的类型：
//
// - "Full"：完整文件的哈希值
//
// - "Comp"：将文件拆分成多个分片，每一个分片计算Hash之后再合并的哈希值
//
// 哈希值：SHA256哈希值，全大写的16进制字符串格式
type FileHash string

const (
	FullHashPrefix      = "Full"
	CompositeHashPrefix = "Comp"
)

func (h *FileHash) GetPrefix() string {
	return string((*h)[:4])
}

func (h *FileHash) GetHash() string {
	return string((*h)[4:])
}

func (h *FileHash) GetHashPrefix(len int) string {
	return string((*h)[4 : 4+len])
}

func (h *FileHash) IsFullHash() bool {
	return (*h)[:4] == FullHashPrefix
}

func (h *FileHash) IsCompositeHash() bool {
	return (*h)[:4] == CompositeHashPrefix
}

func ParseHash(hashStr string) (FileHash, error) {
	if len(hashStr) != 4+64 {
		return "", fmt.Errorf("hash string length should be 4+64, but got %d", len(hashStr))
	}

	prefix := hashStr[:4]
	hash := hashStr[4:]
	if prefix != FullHashPrefix && prefix != CompositeHashPrefix {
		return "", fmt.Errorf("invalid hash prefix: %s", prefix)
	}

	if len(hash) != 64 {
		return "", fmt.Errorf("invalid hash length: %d", len(hash))
	}

	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') {
			return "", fmt.Errorf("invalid hash character: %c", c)
		}
	}

	return FileHash(hashStr), nil
}

func NewFullHash(hash []byte) FileHash {
	return FileHash(FullHashPrefix + strings.ToUpper(hex.EncodeToString(hash)))
}

func CalculateCompositeHash(segmentHashes [][]byte) FileHash {
	data := make([]byte, len(segmentHashes)*32)
	for i, segmentHash := range segmentHashes {
		copy(data[i*32:], segmentHash)
	}
	hash := sha256.Sum256(data)
	return FileHash(CompositeHashPrefix + strings.ToUpper(hex.EncodeToString(hash[:])))
}
