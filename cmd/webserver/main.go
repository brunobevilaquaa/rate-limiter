package main

import "ratelimiter/internal/http"

func main() {
	server := http.NewServer(":8080")

	if err := server.Listen(); err != nil {
		panic(err)
	}
}
