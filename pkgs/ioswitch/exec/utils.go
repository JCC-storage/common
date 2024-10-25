package exec

import (
	"github.com/google/uuid"
	"gitlink.org.cn/cloudream/common/utils/math2"
)

func genRandomPlanID() PlanID {
	return PlanID(uuid.NewString())
}

type Range struct {
	Offset int64
	Length *int64
}

func (r *Range) Extend(other Range) {
	newOffset := math2.Min(r.Offset, other.Offset)

	if r.Length == nil {
		r.Offset = newOffset
		return
	}

	if other.Length == nil {
		r.Offset = newOffset
		r.Length = nil
		return
	}

	otherEnd := other.Offset + *other.Length
	rEnd := r.Offset + *r.Length

	newEnd := math2.Max(otherEnd, rEnd)
	r.Offset = newOffset
	*r.Length = newEnd - newOffset
}

func (r *Range) ExtendStart(start int64) {
	r.Offset = math2.Min(r.Offset, start)
}

func (r *Range) ExtendEnd(end int64) {
	if r.Length == nil {
		return
	}

	rEnd := r.Offset + *r.Length
	newLen := math2.Max(end, rEnd) - r.Offset
	r.Length = &newLen
}

func (r *Range) Fix(maxLength int64) {
	if r.Length != nil {
		return
	}

	len := maxLength - r.Offset
	r.Length = &len
}

func (r *Range) ToStartEnd(maxLen int64) (start int64, end int64) {
	if r.Length == nil {
		return r.Offset, maxLen
	}

	end = r.Offset + *r.Length
	return r.Offset, end
}

func (r *Range) ClampLength(maxLen int64) {
	if r.Length == nil {
		return
	}

	*r.Length = math2.Min(*r.Length, maxLen-r.Offset)
}
