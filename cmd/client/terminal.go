package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type terminalUI struct {
	input  *bufio.Reader
	output io.Writer
}

var ui = terminalUI{
	input:  bufio.NewReader(os.Stdin),
	output: os.Stdout,
}

func (t terminalUI) readLine(prompt string) (string, bool) {
	fmt.Fprint(t.output, prompt)
	line, err := t.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

func (t terminalUI) printf(format string, args ...any) {
	fmt.Fprintf(t.output, format, args...)
}

func (t terminalUI) println(args ...any) {
	fmt.Fprintln(t.output, args...)
}

func (t terminalUI) printResponse(status int, body string) {
	t.printf("HTTP %d %s\n", status, http.StatusText(status))
	if strings.TrimSpace(body) != "" {
		t.println(strings.TrimSpace(body))
	}
}

func (t terminalUI) printMenu(serverURL string) {
	t.printf("\n=== CACHE LAB (%s) ===\n", serverURL)
	t.println("1. Create a key")
	t.println("2. Read a key")
	t.println("3. Update a key")
	t.println("4. Delete a key")
	t.println("5. Test all routes")
	t.println("6. Test hash table resizing (normal mode)")
	t.println("7. Show the OpenAPI schema (/openapi.yaml)")
	t.println("8. Test routes and resizing")
	t.println("9. Run resize milestones (1,000 to 25,000 keys, normal mode)")
	t.println("m. Switch mode (normal / recursive)")
	t.println("s. Show current mode")
	t.println("b. Benchmark every mode on the server (1,000 / 100,000 / 1,000,000 keys)")
	t.printf("0. Quit \n \n \n")
}
