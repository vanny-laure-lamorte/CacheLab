package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/vanny-laure-lamorte/CacheLab/internal/api"
	"github.com/vanny-laure-lamorte/CacheLab/internal/cache"
)

func main() {
	// 1. The cache starts in "normal" mode, the client can switch with PUT /mode
	c, err := cache.New("normal")
	if err != nil {
		fmt.Println("Error creating cache:", err)
		return
	}

	// 2. The routes
	router := api.NewRouter(c)

	// 3. The server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	fmt.Println("Server running on :8080")
	if err := http.Serve(listener, router); err != nil {
		fmt.Println("Error serving HTTP:", err)
	}
}
