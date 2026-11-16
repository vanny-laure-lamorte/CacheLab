// Package cache protects the store against concurrent access
// and lets the client switch between the versions.
package cache

import (
	"sync"

	"github.com/vanny-laure-lamorte/CacheLab/internal/store"
)

// Cache wraps the current store with a read/write lock.
type Cache struct {
	mu    sync.RWMutex
	mode  string
	table store.Store
}

// New creates a cache using the given mode ("normal", "recursive" or "fks").
func New(mode string) (*Cache, error) {
	table, err := store.New(mode)
	if err != nil {
		return nil, err
	}
	return &Cache{mode: mode, table: table}, nil
}

// SetMode replaces the table by an empty one of the given mode.
func (c *Cache) SetMode(mode string) error {
	table, err := store.New(mode) // created outside the lock
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = mode
	c.table = table
	return nil
}

// Mode returns the current mode and its number of keys.
func (c *Cache) Mode() (string, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mode, c.table.Len()
}

// Create adds a new key. Returns false if it already exists.
func (c *Cache) Create(key, value string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.table.CreateKey(key, value)
}

// Get returns the value of a key, and false if it is missing.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.table.GetKey(key)
}

// Update replaces the value of a key. Returns false if it is missing.
func (c *Cache) Update(key, value string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.table.UpdateKey(key, value)
}

// Delete removes a key. Returns false if it is missing.
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.table.DeleteKey(key)
}
