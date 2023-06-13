package reqbuilder

import (
	"gitlink.org.cn/cloudream/common/pkg/distlock"
	mylo "gitlink.org.cn/cloudream/common/utils/lo"
)

type LockRequestBuilder struct {
	locks []distlock.Lock
}

func (b *LockRequestBuilder) Build() distlock.LockRequest {
	return distlock.LockRequest{
		Locks: mylo.ArrayClone(b.locks),
	}
}
