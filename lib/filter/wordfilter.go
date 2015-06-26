package filter

//referer http://blog.csdn.net/chenssy/article/details/26961957

type TrieNode struct {
	words map[rune]*TrieNode
	full  bool
}

func New() *TrieNode {
	t := new(TrieNode)
	t.words = make(map[rune]*TrieNode, 64)
	t.full = false
	return t
}

func (t *TrieNode) Add(world string) {
	node := t
	for _, w := range world {
		_, ok := node.words[w]
		if !ok {
			node.words[w] = &TrieNode{make(map[rune]*TrieNode, 0), false}
		}
		node = node.words[w]
	}
	node.full = true
}

func (t *TrieNode) Filter(world string) string {
	filter := []rune(world)
	length := len(filter)
	matches := make([]int, 0, 4)
	for i := 0; i < length; {
		node := t
		matches = matches[0:0]
		for j := i; j < length; j++ {
			w := filter[j]
			node = node.words[w]
			if node == nil {
				break
			} else {
				matches = append(matches, j)
				if node.full {
					for _, v := range matches {
						filter[v] = '*'
					}
					break
				}
			}
		}
		//skip **
		if len(matches) > 0 {
			i += len(matches)
		} else {
			i++
		}
	}
	return string(filter)
}
