package http

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"ratelimiter/internal/http/handlers"
	"ratelimiter/pkg/ratelimiter"
	"time"
)

type Server struct {
	address string
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
	}
}

func (s *Server) Listen() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handlers.HelloWorldHandler)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	tw, err := time.ParseDuration("10s")
	if err != nil {
		log.Fatal(err)
	}

	rlc := ratelimiter.Config{
		TimeWindow:           tw,
		CreditsPerTimeWindow: 10,
	}

	wrappedMux := ratelimiter.NewUserRateLimiter(mux, rdb, rlc, "")

	log.Println("Server starting on", s.address)
	return http.ListenAndServe(s.address, wrappedMux)
}
