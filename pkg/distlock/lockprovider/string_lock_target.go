package lockprovider

type StringLockTarget struct {
	Components []StringLockTargetComponet
}

// IsConflict 判断两个锁对象是否冲突。注：只有相同的结构的Target才有意义
func (t *StringLockTarget) IsConflict(other *StringLockTarget) bool {
	if len(t.Components) != len(other.Components) {
		return false
	}

	for i := 0; i < len(t.Components); i++ {
		if t.Components[i].IsEquals(&other.Components[i]) {
			return true
		}
	}

	return false
}

type StringLockTargetComponet struct {
	Values []string
}

// IsEquals 判断两个Component是否相同。注：只有相同的结构的Component才有意义
func (t *StringLockTargetComponet) IsEquals(other *StringLockTargetComponet) bool {
	if len(t.Values) != len(other.Values) {
		return false
	}

	for i := 0; i < len(t.Values); i++ {
		if t.Values[i] != other.Values[i] {
			return false
		}
	}

	return true
}
