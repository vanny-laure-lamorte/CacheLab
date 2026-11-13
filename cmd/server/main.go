package main

import (
	"encoding/json"
	"fmt"
	"hash/maphash"
	"net"
	"net/http"
	"sync"

	apidocs "github.com/vanny-laure-lamorte/CacheLab/docs"
)

// maxLoadFactor is the maximum load factor before resizing the hash map.
const maxLoadFactor = 0.75

var (
	mu    sync.RWMutex
	table = NewHashMap(8, true)
)

// pair is the JSON format exchanged with the client.
type pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Entry represents a key-value pair in the hash map, with a pointer to the next entry for handling collisions.
type Entry struct {
	Key   string
	Value string
	Next  *Entry
}

// HashMap is a simple hash map implementation that uses separate chaining to handle collisions.
type HashMap struct {
	buckets []*Entry
	size    int
	seed    maphash.Seed

	resizeEnabled bool
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Server working.")
}

// NewHashMap creates an empty HashMap with the given number of buckets.
func NewHashMap(bucketCount int, resizeEnabled bool) *HashMap {
	if bucketCount < 1 {
		bucketCount = 16
	}
	return &HashMap{
		buckets: make([]*Entry, bucketCount),
		seed:    maphash.MakeSeed(),
		resizeEnabled: resizeEnabled,
	}
}

// hashFunction returns the bucket index of a key, in the range [0, number of buckets).
func (h *HashMap) hashFunction(key string) uint {
	return uint(maphash.String(h.seed, key) % uint64(len(h.buckets)))
}

// resize doubles the number of buckets and rehashes all existing entries.
func (h *HashMap) loadFactor() float64 {
	return float64(h.size) / float64(len(h.buckets))
}

// resize doubles the number of buckets and rehashes all existing entries.
func (h *HashMap) needsResize() bool {
	return h.resizeEnabled && h.loadFactor() > maxLoadFactor
}

// resize doubles the number of buckets and moves every entry to its new bucket.
func (h *HashMap) resize() {
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
}

// CreateKey adds a new pair.
// Returns false if the key already exists (nothing is changed).
func (h *HashMap) CreateKey(key string, value string) bool {
	// Compute the bucket index for the key
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
				h.buckets[index] = entry.Next
			} else {
				prev.Next = entry.Next
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

/*
* writeJSON writes a JSON response to the HTTP response writer.
* It sets the Content-Type header to "application/json", writes the status code,
* and encodes the provided value as JSON.
 */
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

/*
* POST /keys creates a new key-value pair.
* Returns 400 if the JSON is invalid, 409 if the key already exists, 201 otherwise.
 */
func createKeyHandler(w http.ResponseWriter, r *http.Request) {
	var p pair

	// Read the JSON body into p
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide ou clé manquante"})
		return
	}

	mu.Lock()
	ok := table.CreateKey(p.Key, p.Value)
	mu.Unlock()

	if !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "clé déjà existante"})
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

/*
* GET /keys/{key} returns the value associated with the specified key.
* If the key does not exist, it returns a 404 Not Found error.
 */
func getKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	mu.RLock()
	value, ok := table.GetKey(key)
	mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "clé introuvable"})
		return
	}

	writeJSON(w, http.StatusOK, pair{Key: key, Value: value})
}

/*
* PUT /keys/{key} replaces the value of an existing key.
* Returns 400 if the JSON is invalid, 404 if the key does not exist, 200 otherwise.
 */
func updateKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var p pair
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	mu.Lock()
	ok := table.UpdateKey(key, p.Value)
	mu.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "clé introuvable"})
		return
	}
	writeJSON(w, http.StatusOK, pair{Key: key, Value: p.Value})
}

/*
* DELETE /keys/{key} removes a key.
* Returns 404 if the key does not exist, 204 otherwise.
 */
func deleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	mu.Lock()
	ok := table.DeleteKey(key)
	mu.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "clé introuvable"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// docsHandler serves the Swagger UI page, which reads /openapi.yaml.
func docsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
	<html>
	<head>
	<title>CacheLab API</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css">
	</head>
	<body>
	<div id="swagger-ui"></div>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
	<script>
		SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" });
	</script>
	</body>
	</html>`)
}

func main() {
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("POST /keys", createKeyHandler)
	http.HandleFunc("GET /keys/{key}", getKeyHandler)
	http.HandleFunc("PUT /keys/{key}", updateKeyHandler)
	http.HandleFunc("DELETE /keys/{key}", deleteKeyHandler)
	http.HandleFunc("GET /docs", docsHandler)
	http.Handle("GET /openapi.yaml", http.FileServer(http.FS(apidocs.OpenAPI)))

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	fmt.Println("Server running on :8080")
	if err := http.Serve(listener, nil); err != nil {
		fmt.Println("Error serving HTTP:", err)
	}
}
