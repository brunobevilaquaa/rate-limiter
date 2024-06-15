package ratelimiter

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

type MiddlewareConfig struct {
	TimeWindow      time.Duration
	CreditsPerIp    int
	CreditsPerToken int
}

type UserRateLimiter struct {
	handler http.Handler
	cache   Cache
	config  MiddlewareConfig
}

func NewUserRateLimiter(handler http.Handler, cache Cache, creditsPerToken int, creditsPerIp int, timeWindow time.Duration) *UserRateLimiter {
	return &UserRateLimiter{
		handler: handler,
		cache:   cache,
		config: MiddlewareConfig{
			TimeWindow:      timeWindow,
			CreditsPerIp:    creditsPerIp,
			CreditsPerToken: creditsPerToken,
		},
	}
}

func (url *UserRateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	config := url.getConfig(r)

	rl := NewLimiterService(url.cache, config)
	key := url.getUserKey(r)

	if allowed, err := rl.Allowed(key); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	} else if !allowed {
		http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		return
	}

	url.handler.ServeHTTP(w, r)
}

func (url *UserRateLimiter) getConfig(r *http.Request) Config {
	if r.Header.Get("API_KEY") != "" {
		return Config{
			TimeWindow:           url.config.TimeWindow,
			CreditsPerTimeWindow: url.config.CreditsPerToken,
		}
	}

	return Config{
		TimeWindow:           url.config.TimeWindow,
		CreditsPerTimeWindow: url.config.CreditsPerIp,
	}
}

func (url *UserRateLimiter) getClientIP(r *http.Request) string {
	ip := r.RemoteAddr

	ip = strings.Split(ip, ":")[0]

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if !strings.HasPrefix(part, "10.") && !strings.HasPrefix(part, "192.168.") && !strings.HasPrefix(part, "172.") {
				ip = part
				break
			}
		}
	}

	return ip
}

func (url *UserRateLimiter) getUserKey(r *http.Request) string {
	key := r.Header.Get("API_KEY")
	if key == "" {
		key = url.getClientIP(r)
	}

	hash := sha256.New()
	hash.Write([]byte(key))
	hashedKey := hash.Sum(nil)

	return hex.EncodeToString(hashedKey)
}
