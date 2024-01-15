package bitmap

type Bitmap64 uint64

func (b *Bitmap64) Set(index int, val bool) {
	if val {
		*b |= 1 << index
	} else {
		*b &= ^(1 << index)
	}
}

func (b *Bitmap64) Get(index int) bool {
	return (*b & (1 << index)) > 0
}

func (b *Bitmap64) Or(other *Bitmap64) {
	*b |= *other
}

func (b *Bitmap64) Weight() int {
	v := *b
	cnt := 0
	for v > 0 {
		cnt++
		v &= (v - 1)
	}
	return cnt
}
