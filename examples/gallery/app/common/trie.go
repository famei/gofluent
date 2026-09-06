package common

// Trie is a case-insensitive string trie keyed on the lowercase ASCII letters
// a-z (faithful port of app/common/trie.py).
type Trie struct {
	key      string
	value    interface{}
	children []*Trie
	isEnd    bool
}

// NewTrie builds an empty trie.
func NewTrie() *Trie {
	return &Trie{children: make([]*Trie, 26)}
}

func (t *Trie) newChild() *Trie {
	return &Trie{children: make([]*Trie, 26)}
}

// Insert stores value under key. Characters outside a-z stop the insertion
// (matching the Python implementation which simply returns).
func (t *Trie) Insert(key string, value interface{}) {
	node := t
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c < 'a' || c > 'z' {
			return
		}
		idx := int(c - 'a')
		if node.children[idx] == nil {
			node.children[idx] = node.newChild()
		}
		node = node.children[idx]
	}
	node.isEnd = true
	node.key = key
	node.value = value
}

// Get returns the value stored under key, or def when the key is absent.
func (t *Trie) Get(key string, def interface{}) interface{} {
	node := t.searchPrefix(key)
	if node == nil || !node.isEnd {
		return def
	}
	return node.value
}

func (t *Trie) searchPrefix(prefix string) *Trie {
	node := t
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c < 'a' || c > 'z' || node.children[c-'a'] == nil {
			return nil
		}
		node = node.children[c-'a']
	}
	return node
}

// Items returns every (key, value) pair under the given prefix, in
// breadth-first order.
func (t *Trie) Items(prefix string) [][2]interface{} {
	node := t.searchPrefix(prefix)
	if node == nil {
		return nil
	}

	result := make([][2]interface{}, 0)
	queue := []*Trie{node}
	for len(queue) > 0 {
		node = queue[0]
		queue = queue[1:]
		if node.isEnd {
			result = append(result, [2]interface{}{node.key, node.value})
		}
		for _, c := range node.children {
			if c != nil {
				queue = append(queue, c)
			}
		}
	}
	return result
}
