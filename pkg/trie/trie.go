package trie

const (
	WORD_ANY = 0
)

type Node[T any] struct {
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
		node = &Node[T]{}
		n.WordNexts[word] = node
	}

	return node
}

func (n *Node[T]) CreateAny() *Node[T] {
	if n.AnyNext == nil {
		n.AnyNext = &Node[T]{}
	}

	return n.AnyNext
}

type Trie[T any] struct {
	Root Node[T]
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
