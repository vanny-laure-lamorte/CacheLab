package store

import "fmt"

// Store is implemented by HashMap (normal), RecursiveHashMap (recursive).
type Store interface {
	CreateKey(key, value string) bool
	GetKey(key string) (string, bool)
	UpdateKey(key, value string) bool
	DeleteKey(key string) bool
	Len() int
}

// Entry is one key-value pair.
type Entry struct {
	Key   string
	Value string
	Hash  uint64
	Next  *Entry
}

var Logs = true

// Modes lists the available versions.
var Modes = []string{"normal", "recursive"}

// New creates an empty store for the given mode.
func New(mode string) (Store, error) {
	switch mode {
	case "normal":
		return NewHashMap(16, true), nil
	case "recursive":
		return NewRecursiveHashMap(), nil
	default:
		return nil, fmt.Errorf("unknown mode %q", mode)
	}
}
