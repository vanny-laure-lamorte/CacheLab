package api

import (
	"encoding/json"
	"net/http"
)

/*
* GET /mode returns the current version and its number of keys.
 */
func (h *handler) getMode(w http.ResponseWriter, r *http.Request) {
	mode, keys := h.cache.Mode()
	writeJSON(w, http.StatusOK, map[string]any{"mode": mode, "keys": keys})
}

/*
* PUT /mode switches between "normal", "recursive" and.
* The new table starts empty. Returns 400 if the mode is unknown.
 */
func (h *handler) setMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	if err := h.cache.SetMode(body.Mode); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `mode attendu : "normal", "recursive"`})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mode": body.Mode, "keys": 0})
}
