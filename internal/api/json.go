package api

import (
	"encoding/json"
	"net/http"
)

// pair is the JSON format exchanged with the client.
type pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
