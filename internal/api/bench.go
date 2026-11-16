package api

import (
	"net/http"
	"strconv"

	"github.com/vanny-laure-lamorte/CacheLab/internal/bench"
)

// maxBenchSize limits the benchmark size (memory and duration).
const maxBenchSize = 5_000_000

/*
* GET /bench?mode=normal&size=100000 measures one algorithm on the server, without HTTP.
* Returns the time per create and get (ns/op) and the memory used.
* Returns 400 if the mode or the size is invalid.
 */
func (h *handler) runBench(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 || size > maxBenchSize {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "size doit être un entier entre 1 et 5000000"})
		return
	}

	result, err := bench.Run(mode, size)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `mode attendu : "normal" ou "recursive"`})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
