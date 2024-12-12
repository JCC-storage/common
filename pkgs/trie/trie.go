package trie

const (
	WORD_ANY = 0
)

type VisitCtrl int

const (
	VisitContinue = 0
	VisitBreak    = 1
	VisitSkip     = 2
)

type Node[T any] struct {
	Word      any
	Parent    *Node[T]
	WordNexts map[string]*Node[T]
	AnyNext   *Node[T]
	Value     T
}

func (n *Node[T]) WalkNext(word string) *Node[T] {
	if n.WordNexts == nil {
		return n.AnyNext
	}

	node, ok := n.WordNexts[word]
	if ok {
		return node
	}

	return n.AnyNext
}

func (n *Node[T]) walkWordNext(word string) (*Node[T], bool) {
	if n.WordNexts == nil {
		return n.AnyNext, false
	}

	node, ok := n.WordNexts[word]
	if ok {
		return node, true
	}

	return n.AnyNext, false
}

func (n *Node[T]) Create(word string) *Node[T] {
	if n.WordNexts == nil {
		n.WordNexts = make(map[string]*Node[T])
	}

	node, ok := n.WordNexts[word]
	if !ok {
		node = &Node[T]{
			Word:   word,
			Parent: n,
		}
		n.WordNexts[word] = node
	}

	return node
}

func (n *Node[T]) CreateAny() *Node[T] {
	if n.AnyNext == nil {
		n.AnyNext = &Node[T]{
			Word:   WORD_ANY,
			Parent: n,
		}
	}

	return n.AnyNext
}

func (n *Node[T]) IsEmpty() bool {
	return len(n.WordNexts) == 0 && n.AnyNext == nil
}

// 将自己从树中移除。如果cleanParent为true，则会一直向上清除所有没有子节点的节点
func (n *Node[T]) RemoveSelf(cleanParent bool) {
	if n.Parent == nil {
		return
	}

	if n.Word == WORD_ANY {
		if n.Parent.AnyNext == n {
			n.Parent.AnyNext = nil
		}
	} else if n.Parent.WordNexts != nil && n.Parent.WordNexts[n.Word.(string)] == n {
		delete(n.Parent.WordNexts, n.Word.(string))
	}

	if cleanParent {
		if n.Parent.IsEmpty() {
			n.Parent.RemoveSelf(true)
		}
	}

	n.Parent = nil
}

func (n *Node[T]) Iterate(visitorFn func(word string, node *Node[T], isWordNode bool) VisitCtrl) {
	if n.WordNexts != nil {
		for word, node := range n.WordNexts {
			ret := visitorFn(word, node, true)
			if ret == VisitBreak {
				return
			}

			if ret == VisitSkip {
				continue
			}

			node.Iterate(visitorFn)
		}
	}

	if n.AnyNext != nil {
		ret := visitorFn("", n.AnyNext, false)
		if ret == VisitBreak {
			return
		}

		if ret == VisitSkip {
			return
		}

		n.AnyNext.Iterate(visitorFn)
	}
}

type Trie[T any] struct {
	Root Node[T]
}

func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{}
}

func (t *Trie[T]) Walk(words []string, visitorFn func(word string, wordIndex int, node *Node[T], isWordNode bool)) bool {
	ptr := &t.Root

	for index, word := range words {
		var isWord bool
		ptr, isWord = ptr.walkWordNext(word)
		if ptr == nil {
			return false
		}

		visitorFn(word, index, ptr, isWord)
	}

	return true
}

func (t *Trie[T]) WalkEnd(words []string) (*Node[T], bool) {
	ptr := &t.Root

	for _, word := range words {
		ptr = ptr.WalkNext(word)
		if ptr == nil {
			return nil, false
		}
	}

	return ptr, true
}

func (t *Trie[T]) Create(words []any) *Node[T] {
	ptr := &t.Root

	for _, word := range words {
		switch val := word.(type) {
		case string:
			ptr = ptr.Create(val)

		case int:
			ptr = ptr.CreateAny()

		default:
			panic("word can only be string or int 0")
		}
	}

	return ptr
}

func (t *Trie[T]) CreateWords(words []string) *Node[T] {
	ptr := &t.Root

	for _, word := range words {
		ptr = ptr.Create(word)
	}

	return ptr
}

func (n *Trie[T]) Iterate(visitorFn func(word string, node *Node[T], isWordNode bool) VisitCtrl) {
	n.Root.Iterate(visitorFn)
}
