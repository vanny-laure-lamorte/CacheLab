package main

import (
    "fmt"
    "net/http"
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

func main() {

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/about", aboutHandler)

    fmt.Println("Server running on :8080")

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}