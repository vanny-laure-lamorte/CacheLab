package api

import (
	"encoding/json"
	"net/http"
)

/*
* POST /keys creates a new key-value pair.
* Returns 400 if the JSON is invalid, 409 if the key already exists, 201 otherwise.
 */
func (h *handler) createKey(w http.ResponseWriter, r *http.Request) {
	var p pair

	// Read the JSON body into p
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide ou clé manquante"})
		return
	}

	if !h.cache.Create(p.Key, p.Value) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "clé déjà existante"})
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

/*
* GET /keys/{key} returns the value associated with the specified key.
* If the key does not exist, it returns a 404 Not Found error.
 */
func (h *handler) getKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	value, ok := h.cache.Get(key)
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
func (h *handler) updateKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var p pair
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	if !h.cache.Update(key, p.Value) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "clé introuvable"})
		return
	}
	writeJSON(w, http.StatusOK, pair{Key: key, Value: p.Value})
}

/*
* DELETE /keys/{key} removes a key.
* Returns 404 if the key does not exist, 204 otherwise.
 */
func (h *handler) deleteKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if !h.cache.Delete(key) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "clé introuvable"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
