package main

import (
	"encoding/json"
	"fmt"
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
	client   = &http.Client{Timeout: 10 * time.Second}
	inputEOF bool
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

func runResizeTest(count int) {
	if count < 7 {
		ui.println("Choose at least 7 keys to trigger a resize with an empty table.")
		return
	}

	prefix := fmt.Sprintf("client-resize-%d-", time.Now().UnixNano())
	created := make([]string, 0, count)
	start := time.Now()
	failed := 0

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s%d", prefix, i)
		body, err := json.Marshal(keyValue{Key: key, Value: "resize-test"})
		if err != nil {
			ui.printf("Failed to prepare JSON: %v\n", err)
			failed++
			continue
		}
		status, responseBody, err := sendRequest(http.MethodPost, "/keys", string(body))
		if err != nil {
			ui.printf("Failed to create key %s: %v\n", key, err)
			failed++
			continue
		}
		if status != http.StatusCreated {
			ui.printf("Failed to create key %s: HTTP %d %s\n", key, status, strings.TrimSpace(responseBody))
			failed++
			continue
		}
		created = append(created, key)
	}

	for _, key := range created {
		status, _, err := sendRequest(http.MethodDelete, keyRoute(key), "")
		if err != nil {
			ui.printf("Failed to clean up key %s: %v\n", key, err)
			failed++
		} else if status != http.StatusNoContent {
			ui.printf("Failed to clean up key %s: HTTP %d\n", key, status)
			failed++
		}
	}

	ui.printf("Resizing test: created %d/%d keys, %d error(s), cleanup completed in %v.\n",
		len(created), count, failed, time.Since(start).Round(time.Millisecond))
	ui.println(`Check the server terminal for "[HashMap] resized" logs to confirm the table grew.`)
}

func runResizeMilestones() {
	const maxKeys = 25000

	prefix := fmt.Sprintf("client-milestone-%d-", time.Now().UnixNano())
	created := make([]string, 0, maxKeys)
	start := time.Now()
	failed := 0
	nextMilestone := 1000

	for i := 0; i < maxKeys; i++ {
		key := fmt.Sprintf("%s%d", prefix, i)
		body, err := json.Marshal(keyValue{Key: key, Value: "resize-test"})
		if err != nil {
			ui.printf("Failed to prepare JSON: %v\n", err)
			failed++
			continue
		}
		status, responseBody, err := sendRequest(http.MethodPost, "/keys", string(body))
		if err != nil {
			ui.printf("Failed to create key %s: %v\n", key, err)
			failed++
			continue
		}
		if status != http.StatusCreated {
			ui.printf("Failed to create key %s: HTTP %d %s\n", key, status, strings.TrimSpace(responseBody))
			failed++
			continue
		}
		created = append(created, key)

		if len(created) == nextMilestone {
			ui.printf("%5d keys created in %v\n", nextMilestone, time.Since(start).Round(time.Millisecond))
			nextMilestone += 500
		}
	}

	for _, key := range created {
		status, _, err := sendRequest(http.MethodDelete, keyRoute(key), "")
		if err != nil {
			ui.printf("Failed to clean up key %s: %v\n", key, err)
			failed++
		} else if status != http.StatusNoContent {
			ui.printf("Failed to clean up key %s: HTTP %d\n", key, status)
			failed++
		}
	}

	ui.printf("Resize milestones: created and cleaned up %d/%d keys, %d error(s), completed in %v.\n",
		len(created), maxKeys, failed, time.Since(start).Round(time.Millisecond))
	ui.println(`Check the server terminal for "[HashMap] resized" logs to confirm the table grew.`)
}

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
