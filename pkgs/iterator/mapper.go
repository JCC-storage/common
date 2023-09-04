package iterator

type Mapper[TSrc any, TDst any] struct {
	srcIter Iterator[TSrc]
	mapper  func(TSrc) (TDst, error)
}

func (i *Mapper[TSrc, TDst]) MoveNext() (TDst, error) {
	src, err := i.srcIter.MoveNext()
	if err != nil {
		var ret TDst
		return ret, err
	}

	return i.mapper(src)
}

func (i *Mapper[TSrc, TDst]) Close() {
	i.srcIter.Close()
}

func Map[TSrc any, TDst any](srcIter Iterator[TSrc], mapper func(src TSrc) (TDst, error)) *Mapper[TSrc, TDst] {
	return &Mapper[TSrc, TDst]{
		srcIter: srcIter,
		mapper:  mapper,
	}
}
