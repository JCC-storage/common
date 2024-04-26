package sync2

import (
	"reflect"

	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type SelectCase int

type SelectSet[T any, C any] struct {
	cases []reflect.SelectCase
	tags  []T
}

func (s *SelectSet[T, C]) Add(tag T, ch <-chan C) SelectCase {
	s.cases = append(s.cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
	s.tags = append(s.tags, tag)

	return SelectCase(len(s.cases) - 1)
}

func (s *SelectSet[T, C]) AddDefault(tag T, ch <-chan C) SelectCase {
	s.cases = append(s.cases, reflect.SelectCase{Dir: reflect.SelectDefault, Chan: reflect.ValueOf(ch)})
	s.tags = append(s.tags, tag)

	return SelectCase(len(s.cases) - 1)
}

func (s *SelectSet[T, C]) Remove(caze SelectCase) {
	s.cases = lo2.RemoveAt(s.cases, int(caze))
	s.tags = lo2.RemoveAt(s.tags, int(caze))
}

func (s *SelectSet[T, C]) Select() (T, C, bool) {
	chosen, recv, ok := reflect.Select(s.cases)
	if !ok {
		var t T
		var c C
		return t, c, false
	}

	return s.tags[chosen], recv.Interface().(C), true
}

func (s *SelectSet[T, C]) Count() int {
	return len(s.cases)
}
