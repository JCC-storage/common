package math2

type Range struct {
	Offset int64
	Length *int64
}

// length为-1时Range.Length为nil
func NewRange(offset int64, length int64) Range {
	if length == -1 {
		return Range{Offset: offset, Length: nil}
	}
	return Range{Offset: offset, Length: &length}
}

// 不包含end
func RangeFromStartEnd(start int64, end int64) Range {
	length := end - start
	return Range{Offset: start, Length: &length}
}

// 给Length设置一个具体值
func (r *Range) Fix(totalLen int64) {
	len := totalLen - r.Offset
	r.Length = &len
}

// 如果Length为nil，则end为-1
func (r *Range) ToStartEnd() (start int64, end int64) {
	if r.Length == nil {
		return r.Offset, -1
	}

	end = r.Offset + *r.Length
	return r.Offset, end
}

// 将范围限制在totalLen内。会同时设置Length的值
func (r *Range) Clamp(totalLen int64) {
	r.Offset = Min(r.Offset, totalLen)
	if r.Length == nil {
		len := totalLen - r.Offset
		r.Length = &len
	} else {
		*r.Length = Min(*r.Length, totalLen-r.Offset)
	}
}

func (r *Range) Extend(other Range) {
	newOffset := Min(r.Offset, other.Offset)

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

	newEnd := Max(otherEnd, rEnd)
	r.Offset = newOffset
	*r.Length = newEnd - newOffset
}

func (r *Range) ExtendStart(start int64) {
	r.Offset = Min(r.Offset, start)
}

func (r *Range) ExtendEnd(end int64) {
	if r.Length == nil {
		return
	}

	rEnd := r.Offset + *r.Length
	newLen := Max(end, rEnd) - r.Offset
	r.Length = &newLen
}

func (r *Range) Equals(other Range) bool {
	if r.Offset != other.Offset {
		return false
	}

	if r.Length == nil && other.Length == nil {
		return true
	}

	if r.Length == nil || other.Length == nil {
		return false
	}

	return *r.Length == *other.Length
}
