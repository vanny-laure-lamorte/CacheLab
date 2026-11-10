package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	apidocs "github.com/vanny-laure-lamorte/CacheLab/docs"
)

type pair struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

var (
    mu   sync.RWMutex
    data = make(map[string]string)
)

/*
* This is a simple Go web server that responds to HTTP requests.
* It has two routes: the home page ("/") and the about page ("/about").
* The server listens on port 8080.
*/
func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello from Go!")
}

/*
* This function handles requests to the "/about" route.
* It responds with a simple message indicating that this is the about page.
*/
func aboutHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "This is the about page.")
}

/*
* This function writes a JSON response to the HTTP response writer.
* It sets the Content-Type header to "application/json", writes the status code,
* and encodes the provided value as JSON.
*/
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

/*
* This function handles POST requests to create a new key-value pair.
*/
func createKeyHandler(w http.ResponseWriter, r *http.Request) {
	var p pair
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil || p.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide ou clé manquante"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := data[p.Key]; exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "la clé existe déjà"})
		return
	}

	data[p.Key] = p.Value
	writeJSON(w, http.StatusCreated, p)
}

/*
* GET /keys/{key} returns the value associated with the specified key.
* If the key does not exist, it returns a 404 Not Found error.
*/
func getKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	mu.RLock()
	value, ok := data[key]
	mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "clé introuvable"})
		return
	}

	writeJSON(w, http.StatusOK, pair{Key: key, Value: value})
}

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

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/about", aboutHandler)
    http.HandleFunc("POST /keys", createKeyHandler)
	http.HandleFunc("GET /keys/{key}", getKeyHandler)
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