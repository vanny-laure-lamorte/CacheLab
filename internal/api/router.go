// Package api turns HTTP requests into cache calls.
package api

import (
	"net/http"

	apidocs "github.com/vanny-laure-lamorte/CacheLab/docs"
	"github.com/vanny-laure-lamorte/CacheLab/internal/cache"
)

// handler gives every route access to the cache.
type handler struct {
	cache *cache.Cache
}

// NewRouter declares all the routes in one place.
func NewRouter(c *cache.Cache) *http.ServeMux {
	h := &handler{cache: c}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /mode", h.getMode)
	mux.HandleFunc("PUT /mode", h.setMode)

	mux.HandleFunc("GET /bench", h.runBench)

	mux.HandleFunc("POST /keys", h.createKey)
	mux.HandleFunc("GET /keys/{key}", h.getKey)
	mux.HandleFunc("PUT /keys/{key}", h.updateKey)
	mux.HandleFunc("DELETE /keys/{key}", h.deleteKey)

	mux.HandleFunc("GET /docs", h.swagger)
	mux.Handle("GET /openapi.yaml", http.FileServer(http.FS(apidocs.OpenAPI)))

	return mux
}
