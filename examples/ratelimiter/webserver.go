package main

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"ratelimiter/pkg/ratelimiter"
	"strconv"
	"time"
)

func helloWorldHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Hello, World!"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloWorldHandler)

	var timeWindow time.Duration
	var creditsPerTimeWindow int
	var jwtSecret string

	timeWindow, err := time.ParseDuration(os.Getenv("RATE_LIMITER_TIME_WINDOW"))
	if err != nil {
		timeWindow = 10 * time.Second
	}

	creditsPerTimeWindow, err = strconv.Atoi(os.Getenv("RATE_LIMITER_CREDITS_PER_TIME_WINDOW"))
	if err != nil {
		creditsPerTimeWindow = 10
	}

	jwtSecret = os.Getenv("RATE_LIMITER_JWT_SECRET")

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST_ADDR"),
		Password: "",
		DB:       0,
	})

	cache := ratelimiter.NewCacheRepository(rdb)

	rlc := ratelimiter.Config{
		TimeWindow:           timeWindow,
		CreditsPerTimeWindow: creditsPerTimeWindow,
	}

	wrappedMux := ratelimiter.NewUserRateLimiter(mux, cache, rlc, jwtSecret)

	log.Printf("Middleware config, TimeWindow: %s CreditsPerTimeWindow %d", timeWindow, creditsPerTimeWindow)
	log.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}
