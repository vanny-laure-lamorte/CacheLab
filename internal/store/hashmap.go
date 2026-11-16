package store

import (
	"fmt"
	"hash/maphash"
	"time"
)

// maxLoadFactor is the fill rate (keys / buckets) above which the table grows.
const maxLoadFactor = 0.75

// HashMap is a hash map that uses separate chaining (linked lists) to handle collisions.
type HashMap struct {
	buckets []*Entry
	size    int
	seed    maphash.Seed

	resizeEnabled bool
}

// NewHashMap creates an empty HashMap with the given number of buckets.
func NewHashMap(bucketCount int, resizeEnabled bool) *HashMap {
	if bucketCount < 1 {
		bucketCount = 16
	}
	if Logs {
		fmt.Printf("[HashMap] created: %d buckets\n", bucketCount)
	}
	return &HashMap{
		buckets:       make([]*Entry, bucketCount),
		seed:          maphash.MakeSeed(),
		resizeEnabled: resizeEnabled,
	}
}

// hashFunction returns the bucket index of a key, in the range [0, number of buckets).
func (h *HashMap) hashFunction(key string) uint {
	return uint(maphash.String(h.seed, key) % uint64(len(h.buckets)))
}

// loadFactor returns the fill rate of the table: keys / buckets.
func (h *HashMap) loadFactor() float64 {
	return float64(h.size) / float64(len(h.buckets))
}

// needsResize tells if the table must grow: resize enabled and too full.
func (h *HashMap) needsResize() bool {
	return h.resizeEnabled && h.loadFactor() > maxLoadFactor
}

// resize doubles the number of buckets and moves every entry to its new bucket.
func (h *HashMap) resize() {
	start := time.Now()
	oldBuckets := h.buckets
	h.buckets = make([]*Entry, len(oldBuckets)*2)

	for _, entry := range oldBuckets {
		for entry != nil {
			next := entry.Next

			index := h.hashFunction(entry.Key)
			entry.Next = h.buckets[index]
			h.buckets[index] = entry

			entry = next
		}
	}
	if Logs {
		fmt.Printf("[HashMap] resized: %d -> %d buckets, %d keys moved in %d ns\n",
			len(oldBuckets), len(h.buckets), h.size, time.Since(start).Nanoseconds())
	}
}

// CreateKey adds a new pair.
// Returns false if the key already exists (nothing is changed).
func (h *HashMap) CreateKey(key string, value string) bool {
	index := h.hashFunction(key)

	// Check if the key already exists in the bucket
	for entry := h.buckets[index]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			return false
		}
	}

	// Insert the new entry at the beginning of the linked list in the bucket
	h.buckets[index] = &Entry{Key: key, Value: value, Next: h.buckets[index]}
	h.size++

	if h.needsResize() {
		h.resize()
	}
	return true
}

// GetKey returns the value and true, or "" and false if the key is missing.
func (h *HashMap) GetKey(key string) (string, bool) {
	for entry := h.buckets[h.hashFunction(key)]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			return entry.Value, true
		}
	}
	return "", false
}

// UpdateKey replaces the value of an existing key.
// Returns false if the key does not exist (nothing is created).
func (h *HashMap) UpdateKey(key string, value string) bool {
	for entry := h.buckets[h.hashFunction(key)]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			entry.Value = value
			return true
		}
	}
	return false
}

// DeleteKey removes a key.
// Returns false if the key does not exist.
func (h *HashMap) DeleteKey(key string) bool {
	index := h.hashFunction(key)
	var prev *Entry

	for entry := h.buckets[index]; entry != nil; entry = entry.Next {
		if entry.Key == key {
			if prev == nil {
				h.buckets[index] = entry.Next // first of the chain
			} else {
				prev.Next = entry.Next // skip it
			}
			h.size--
			return true
		}
		prev = entry
	}
	return false
}

// Len returns the number of stored keys.
func (h *HashMap) Len() int {
	return h.size
}
