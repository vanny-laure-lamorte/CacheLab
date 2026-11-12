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

func checkResponseStatus(got, want int, ok, total *int) {
	*total++
	if got == want {
		*ok++
	}
}

func runScenario(n int) {
	ok, total := 0, 0
	start := time.Now()

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("test-%d-%d", n, i)
		path := "/keys/" + key

		checkResponseStatus(sendRequest("POST", "/keys", fmt.Sprintf(`{"key":"%s","value":"v%d"}`, key, i)), 201, &ok, &total)
		checkResponseStatus(sendRequest("GET", path, ""), 200, &ok, &total)
		checkResponseStatus(sendRequest("PUT", path, `{"value":"updated"}`), 200, &ok, &total)
		checkResponseStatus(sendRequest("DELETE", path, ""), 204, &ok, &total)
		checkResponseStatus(sendRequest("GET", path, ""), 404, &ok, &total)
	}

	elapsed := time.Since(start)
	fmt.Printf("%4d clés : %4d/%4d requêtes OK en %v\n", n, ok, total, elapsed.Round(time.Millisecond))
}

func main() {

	for _, n := range []int{1, 10, 100, 1000} {
		runScenario(n)
	}
}
