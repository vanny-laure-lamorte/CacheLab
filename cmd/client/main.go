package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Define the routes
	routes := []string{
		"/about",
	}

	// Iterate over the routes and make GET requests
	for _, route := range routes {
		resp, err := client.Get("http://localhost:8080" + route)
		if err != nil {
			fmt.Println("[ERROR] Erreur :", err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Println("[ERROR] Erreur de lecture :", err)
			return
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("[SUCCESS] GET %s -> %d : %s\n",
				route, resp.StatusCode, body)
		} else {
			fmt.Printf("[WARNING] GET %s -> %d : %s\n",
				route, resp.StatusCode, body)
		}
	}
}
