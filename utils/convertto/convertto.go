package convertto

import (
	"github.com/inhies/go-bytesize"
)

func GBToBytes(gb float64) int64 {
	// 将 float64 转换为 bytesize.ByteSize 类型
	size := bytesize.GB * bytesize.ByteSize(gb)

	// 获取字节数
	bytes := int64(size)

	return bytes
}

func BytesToGB(bytes int64) float64 {
	// 将字节数转换成 ByteSize 类型
	size := bytesize.B * bytesize.ByteSize(bytes)

	// 获取 GB 值
	gb := float64(size)

	return gb

}
