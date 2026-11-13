package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const baseURL = "http://localhost:8080"

var client = &http.Client{
	Timeout: 5 * time.Second,
}

/*
* sendRequest sends an HTTP request to the server with the specified method, route, and body.
* It returns the HTTP status code of the response.
 */
func sendRequest(method, route, body string) int {
	req, err := http.NewRequest(method, baseURL+route, strings.NewReader(body))
	if err != nil {
		return 0
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}

	resp.Body.Close()
	return resp.StatusCode
}

/*
* checkResponseStatus checks if the received HTTP status code matches the expected status code.
* It increments the total request count and the successful request count if they match.
 */
func checkResponseStatus(got, want int, ok, total *int) {
	*total++
	if got == want {
		*ok++
	}
}

/*
* runScenario runs a series of HTTP requests to test the key-value store.
* It creates, retrieves, updates, and deletes keys, checking the response status codes.
* It measures the time taken to complete all requests and prints the results.
 */
func runScenario(n int) {
	ok, total := 0, 0
	start := time.Now()

	// 1. Create all keys (the table fills up → resize)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("test-%d-%d", n, i)
		checkResponseStatus(sendRequest("POST", "/keys", fmt.Sprintf(`{"key":"%s","value":"v%d"}`, key, i)), 201, &ok, &total)
	}

	// 2. Read, update, delete, check deletion
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("test-%d-%d", n, i)
		path := "/keys/" + key

		checkResponseStatus(sendRequest("GET", path, ""), 200, &ok, &total)
		checkResponseStatus(sendRequest("PUT", path, `{"value":"updated"}`), 200, &ok, &total)
		checkResponseStatus(sendRequest("DELETE", path, ""), 204, &ok, &total)
		checkResponseStatus(sendRequest("GET", path, ""), 404, &ok, &total)
	}

	elapsed := time.Since(start)
	fmt.Printf("%4d clés : %4d/%4d requêtes OK en %v\n", n, ok, total, elapsed.Round(time.Millisecond))
}

/*
* main is the entry point of the client application.
* It runs the test scenarios for different numbers of keys.
 */
func main() {

	for _, n := range []int{1000, 5000, 10000, 15000, 25000} {
		runScenario(n)
	}
}
