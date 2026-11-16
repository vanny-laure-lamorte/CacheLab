package store

import (
	"fmt"
	"hash/maphash"
)

const (
	bitsPerLevel = 5
	width        = 1 << bitsPerLevel
	mask         = width - 1
)

// node is one level: each slot holds nothing, one entry, or a child node.
type node struct {
	entries  [width]*Entry
	children [width]*node
}

// RecursiveHashMap splits a slot into a child node when two keys collide
type RecursiveHashMap struct {
	root *node
	size int
	seed maphash.Seed
}

// NewRecursiveHashMap creates an empty recursive map (no size to choose, no resize).
func NewRecursiveHashMap() *RecursiveHashMap {
	if Logs {
		fmt.Println("[RecursiveHashMap] created")
	}
	return &RecursiveHashMap{root: &node{}, seed: maphash.MakeSeed()}
}

// levelIndex returns the slot of a hash at a given level (5 different bits per level).
func levelIndex(hash uint64, level int) int {
	return int((hash >> (bitsPerLevel * level)) & mask)
}

// CreateKey adds a new pair. Returns false if the key already exists.
func (m *RecursiveHashMap) CreateKey(key, value string) bool {
	e := &Entry{Key: key, Value: value, Hash: maphash.String(m.seed, key)}
	if insert(m.root, e, 0) {
		m.size++
		return true
	}
	return false
}

func insert(n *node, e *Entry, level int) bool {
	i := levelIndex(e.Hash, level)

	// 1. A child node: go down
	if child := n.children[i]; child != nil {
		return insert(child, e, level+1)
	}

	cur := n.entries[i]

	// 2. Empty slot: store here
	if cur == nil {
		n.entries[i] = e
		return true
	}

	// 3. Same key: refuse
	for c := cur; c != nil; c = c.Next {
		if c.Key == e.Key {
			return false
		}
	}

	// 4. Same full hash: impossible to split, chain them
	if cur.Hash == e.Hash {
		e.Next = cur
		n.entries[i] = e
		return true
	}

	// 5. Collision: replace the entry by a child node, put both inside
	child := &node{}
	n.entries[i] = nil
	n.children[i] = child
	insert(child, cur, level+1)
	return insert(child, e, level+1)
}

// leaf goes down the levels and returns the node and slot where a key would be.
func (m *RecursiveHashMap) leaf(key string) (*node, int) {
	hash := maphash.String(m.seed, key)
	n, level := m.root, 0

	// Go down while the slot holds a child node
	for n.children[levelIndex(hash, level)] != nil {
		n = n.children[levelIndex(hash, level)]
		level++
	}
	return n, levelIndex(hash, level)
}

// find returns the entry of a key, or nil.
func (m *RecursiveHashMap) find(key string) *Entry {
	n, i := m.leaf(key)
	for e := n.entries[i]; e != nil; e = e.Next {
		if e.Key == key {
			return e
		}
	}
	return nil
}

// GetKey returns the value and true, or "" and false if the key is missing.
func (m *RecursiveHashMap) GetKey(key string) (string, bool) {
	if e := m.find(key); e != nil {
		return e.Value, true
	}
	return "", false
}

// UpdateKey replaces the value of an existing key. Returns false if missing.
func (m *RecursiveHashMap) UpdateKey(key, value string) bool {
	if e := m.find(key); e != nil {
		e.Value = value
		return true
	}
	return false
}

// DeleteKey removes a key. Returns false if missing.
// Empty child nodes are kept.
func (m *RecursiveHashMap) DeleteKey(key string) bool {
	n, i := m.leaf(key)
	var prev *Entry

	for e := n.entries[i]; e != nil; e = e.Next {
		if e.Key == key {
			if prev == nil {
				n.entries[i] = e.Next
			} else {
				prev.Next = e.Next
			}
			m.size--
			return true
		}
		prev = e
	}
	return false
}

// Len returns the number of stored keys.
func (m *RecursiveHashMap) Len() int {
	return m.size
}
