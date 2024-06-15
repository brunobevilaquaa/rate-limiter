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
	var creditsPerToken int
	var creditsPerIp int

	timeWindow, err := time.ParseDuration(os.Getenv("RATE_LIMITER_TIME_WINDOW"))
	if err != nil {
		panic(err)
	}

	creditsPerToken, err = strconv.Atoi(os.Getenv("RATE_LIMITER_CREDITS_PER_TOKEN"))
	if err != nil {
		panic(err)
	}

	creditsPerIp, err = strconv.Atoi(os.Getenv("RATE_LIMITER_CREDITS_PER_IP"))
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST_ADDR"),
		Password: "",
		DB:       0,
	})

	cache := ratelimiter.NewCacheRepository(rdb)

	wrappedMux := ratelimiter.NewUserRateLimiter(mux, cache, creditsPerToken, creditsPerIp, timeWindow)

	log.Printf("Middleware config, TimeWindow: %s CreditsPerToken %d CreditsPerToken %d", timeWindow, creditsPerToken, creditsPerIp)
	log.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}
