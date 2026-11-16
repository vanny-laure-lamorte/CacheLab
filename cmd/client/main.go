package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "http://localhost:8080"

var (
	baseURL  = defaultBaseURL
	client   = &http.Client{Timeout: 2 * time.Minute} // long enough for /bench
	inputEOF bool
	modes    = []string{"normal", "recursive"}
)

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func sendRequest(method, route, body string) (int, string, error) {
	req, err := http.NewRequest(method, baseURL+route, strings.NewReader(body))
	if err != nil {
		return 0, "", fmt.Errorf("create request: %w", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("connect to server %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("read response: %w", err)
	}
	return resp.StatusCode, string(responseBody), nil
}

func readLine(prompt string) string {
	line, eof := ui.readLine(prompt)
	inputEOF = eof
	return line
}

func requestAndPrint(method, route, body string) {
	status, responseBody, err := sendRequest(method, route, body)
	if err != nil {
		ui.printf("Error: %v\n", err)
		return
	}
	ui.printResponse(status, responseBody)
}

func keyRoute(key string) string {
	return "/keys/" + url.PathEscape(key)
}

// showMode prints the current version and its number of keys.
func showMode() {
	requestAndPrint(http.MethodGet, "/mode", "")
}

// setMode switches the server to "normal" or "recursive" (the table starts empty).
func setMode(mode string) bool {
	body, err := json.Marshal(map[string]string{"mode": mode})
	if err != nil {
		ui.printf("Failed to prepare JSON: %v\n", err)
		return false
	}
	status, responseBody, err := sendRequest(http.MethodPut, "/mode", string(body))
	if err != nil {
		ui.printf("Failed to switch to %q: %v\n", mode, err)
		return false
	}
	if status != http.StatusOK {
		ui.printf("Failed to switch to %q: HTTP %d %s\n", mode, status, strings.TrimSpace(responseBody))
		return false
	}
	ui.printf("Mode: %s (empty table)\n", mode)
	return true
}

// Manual CRUD

func createKey() {
	key := readLine("Key: ")
	if key == "" {
		ui.println("Key cannot be empty.")
		return
	}
	value := readLine("Value: ")
	body, err := json.Marshal(keyValue{Key: key, Value: value})
	if err != nil {
		ui.printf("Failed to prepare JSON: %v\n", err)
		return
	}
	requestAndPrint(http.MethodPost, "/keys", string(body))
}

func getKey() {
	key := readLine("Key to read: ")
	if key == "" {
		ui.println("Key cannot be empty.")
		return
	}
	requestAndPrint(http.MethodGet, keyRoute(key), "")
}

func updateKey() {
	key := readLine("Key to update: ")
	if key == "" {
		ui.println("Key cannot be empty.")
		return
	}
	value := readLine("New value: ")
	body, err := json.Marshal(map[string]string{"value": value})
	if err != nil {
		ui.printf("Failed to prepare JSON: %v\n", err)
		return
	}
	requestAndPrint(http.MethodPut, keyRoute(key), string(body))
}

func deleteKey() {
	key := readLine("Key to delete: ")
	if key == "" {
		ui.println("Key cannot be empty.")
		return
	}
	requestAndPrint(http.MethodDelete, keyRoute(key), "")
}

type routeCheck struct {
	name   string
	method string
	route  string
	body   string
	want   int
}

func runRouteChecks() {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	key := "client-test-" + suffix
	missingKey := key + "-missing"
	createBody, err := json.Marshal(keyValue{Key: key, Value: "initial"})
	if err != nil {
		ui.printf("Failed to prepare route test: %v\n", err)
		return
	}

	checks := []routeCheck{
		{name: "POST /keys (create)", method: http.MethodPost, route: "/keys", body: string(createBody), want: http.StatusCreated},
		{name: "POST /keys (duplicate key)", method: http.MethodPost, route: "/keys", body: string(createBody), want: http.StatusConflict},
		{name: "POST /keys (invalid JSON)", method: http.MethodPost, route: "/keys", body: `{"key":`, want: http.StatusBadRequest},
		{name: "GET /keys/{key}", method: http.MethodGet, route: keyRoute(key), want: http.StatusOK},
		{name: "GET /keys/{key} (missing key)", method: http.MethodGet, route: keyRoute(missingKey), want: http.StatusNotFound},
		{name: "PUT /keys/{key}", method: http.MethodPut, route: keyRoute(key), body: `{"value":"updated"}`, want: http.StatusOK},
		{name: "PUT /keys/{key} (invalid JSON)", method: http.MethodPut, route: keyRoute(key), body: `{"value":`, want: http.StatusBadRequest},
		{name: "PUT /keys/{key} (missing key)", method: http.MethodPut, route: keyRoute(missingKey), body: `{"value":"x"}`, want: http.StatusNotFound},
		{name: "DELETE /keys/{key}", method: http.MethodDelete, route: keyRoute(key), want: http.StatusNoContent},
		{name: "DELETE /keys/{key} (missing key)", method: http.MethodDelete, route: keyRoute(missingKey), want: http.StatusNotFound},
		{name: "GET /mode", method: http.MethodGet, route: "/mode", want: http.StatusOK},
		{name: "PUT /mode (unknown mode)", method: http.MethodPut, route: "/mode", body: `{"mode":"x"}`, want: http.StatusBadRequest},
		{name: "GET /docs", method: http.MethodGet, route: "/docs", want: http.StatusOK},
		{name: "GET /openapi.yaml", method: http.MethodGet, route: "/openapi.yaml", want: http.StatusOK},
	}

	ok, total := 0, len(checks)
	start := time.Now()
	for _, check := range checks {
		status, _, err := sendRequest(check.method, check.route, check.body)
		if err != nil {
			ui.printf("FAIL   %-43s %v\n", check.name, err)
			continue
		}
		if status == check.want {
			ok++
			ui.printf("PASS   %-43s HTTP %d\n", check.name, status)
		} else {
			ui.printf("FAIL   %-43s expected HTTP %d, got HTTP %d\n", check.name, check.want, status)
		}
	}
	ui.printf("Routes: %d/%d checks passed in %v\n", ok, total, time.Since(start).Round(time.Millisecond))
}

// benchKey returns the name of the i-th test key.
func benchKey(i int) string {
	return fmt.Sprintf("bench-%d", i)
}

// createKeys creates count keys (bench-0, bench-1...) and returns the number of errors.
// onCreated is called after each key (can be nil).
func createKeys(count int, onCreated func(created int)) int {
	failed := 0
	for i := 0; i < count; i++ {
		body, err := json.Marshal(keyValue{Key: benchKey(i), Value: "v"})
		if err != nil {
			failed++
			continue
		}
		status, _, err := sendRequest(http.MethodPost, "/keys", string(body))
		if err != nil || status != http.StatusCreated {
			failed++
			continue
		}
		if onCreated != nil {
			onCreated(i + 1)
		}
	}
	return failed
}

func runResizeTest(count int) {
	if !setMode("normal") { // empty table, so the resizes happen again
		return
	}
	start := time.Now()
	failed := createKeys(count, nil)

	ui.printf("Resizing test: created %d/%d keys, %d error(s) in %v.\n",
		count-failed, count, failed, time.Since(start).Round(time.Millisecond))
	ui.println(`Check the server terminal for "[HashMap] resized" logs to confirm the table grew.`)
}

func runResizeMilestones() {
	const maxKeys = 25000

	if !setMode("normal") {
		return
	}
	start := time.Now()
	nextMilestone := 1000

	failed := createKeys(maxKeys, func(created int) {
		if created == nextMilestone {
			ui.printf("%5d keys created in %v\n", created, time.Since(start).Round(time.Millisecond))
			nextMilestone += 500
		}
	})

	ui.printf("Resize milestones: created %d/%d keys, %d error(s), completed in %v.\n",
		maxKeys-failed, maxKeys, failed, time.Since(start).Round(time.Millisecond))
	ui.println(`Check the server terminal for "[HashMap] resized" logs to confirm the table grew.`)
}

type benchResult struct {
	Mode          string  `json:"mode"`
	Size          int     `json:"size"`
	CreateNsPerOp float64 `json:"create_ns_per_op"`
	GetNsPerOp    float64 `json:"get_ns_per_op"`
	CreateTotalMs float64 `json:"create_total_ms"`
	GetTotalMs    float64 `json:"get_total_ms"`
	MemoryMB      float64 `json:"memory_mb"`
	DurationMs    float64 `json:"duration_ms"`
}

// runServerBenchmark asks the server to measure every mode for every size.
func runServerBenchmark() {
	ui.println("Server benchmark (median of 3 runs, no HTTP during the measure):")
	ui.printf("%-9s %9s | %12s | %12s | %10s | %10s | %10s\n",
		"mode", "keys", "create ns/op", "create total", "get ns/op", "get total", "memory")
	totalMs := 0.0
	for _, size := range []int{1000, 100000, 1000000} {
		for _, mode := range modes {
			route := fmt.Sprintf("/bench?mode=%s&size=%d", url.QueryEscape(mode), size)
			status, body, err := sendRequest(http.MethodGet, route, "")
			if err != nil {
				ui.printf("Error: %v\n", err)
				return
			}
			if status != http.StatusOK {
				ui.printf("HTTP %d %s\n", status, strings.TrimSpace(body))
				return
			}

			var r benchResult
			if err := json.Unmarshal([]byte(body), &r); err != nil {
				ui.printf("Invalid response: %v\n", err)
				return
			}
			ui.printf("%-9s %9d | %12.1f | %9.3f ms | %10.1f | %7.3f ms | %7.1f Mo\n",
				r.Mode, r.Size, r.CreateNsPerOp, r.CreateTotalMs, r.GetNsPerOp, r.GetTotalMs, r.MemoryMB)
			totalMs += r.DurationMs
		}
	}
	ui.printf("Benchmark completed in %.1f s (server time)\n", totalMs/1000)
}

/*
* main is the entry point of the client application.
* It shows an interactive menu to test the API and compare the versions.
 */
func main() {
	if configuredURL := os.Getenv("CACHELAB_URL"); configuredURL != "" {
		baseURL = strings.TrimRight(configuredURL, "/")
	}

	for {
		ui.printMenu(baseURL)
		switch readLine("Choose an option: ") {
		case "1":
			createKey()
		case "2":
			getKey()
		case "3":
			updateKey()
		case "4":
			deleteKey()
		case "5":
			runRouteChecks()
		case "6":
			countInput := readLine("Number of keys to create: ")
			count, err := strconv.Atoi(countInput)
			if err != nil || count < 7 {
				ui.println("Enter an integer of at least 7.")
				continue
			}
			runResizeTest(count)
		case "7":
			requestAndPrint(http.MethodGet, "/openapi.yaml", "")
		case "8":
			runRouteChecks()
			runResizeTest(1000)
		case "9":
			runResizeMilestones()
		case "m":
			setMode(readLine("Mode (normal / recursive): "))
		case "s":
			showMode()
		case "b":
			runServerBenchmark()
		case "0":
			ui.println("Goodbye.")
			return
		case "":
			if inputEOF {
				ui.println("\nEnd of input.")
				return
			}
			ui.println("Invalid option.")
		default:
			ui.println("Invalid option.")
		}
	}
}
