package ratelimiter

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type MiddlewareConfig struct {
	JWTSecret string
	Config
}

type UserRateLimiter struct {
	handler     http.Handler
	redisClient *redis.Client
	config      Config
	jwtSecret   string
}

func NewUserRateLimiter(handler http.Handler, redisClient *redis.Client, config Config, jwtSecret string) *UserRateLimiter {
	return &UserRateLimiter{
		handler:     handler,
		redisClient: redisClient,
		config:      config,
		jwtSecret:   jwtSecret,
	}
}

func (url *UserRateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cacheRepository := NewCacheRepository(url.redisClient)

	config, err := url.getConfigFromJWTToken(r)
	if err != nil {
		config = &url.config
	}

	rl := NewLimiterService(cacheRepository, *config)
	key := url.getUserKey(r)

	if allowed, err := rl.Allowed(key); err != nil {
		log.Println("error checking if allowed:", err)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	} else if !allowed {
		http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		return
	}

	url.handler.ServeHTTP(w, r)
}

func (url *UserRateLimiter) getConfigFromJWTToken(r *http.Request) (*Config, error) {
	token := r.Header.Get("API_KEY")
	if token == "" {
		return nil, errors.New("missing API_KEY header")
	}

	data, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(url.jwtSecret), nil
	})

	if err != nil {
		return nil, errors.New("error parsing JWT token")
	}

	if !data.Valid {
		return nil, errors.New("invalid JWT token")
	}

	claims, ok := data.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("error parsing JWT claims")
	}

	timeWindow, err := time.ParseDuration(claims["rateLimiterTimeWindow"].(string))
	if err != nil {
		return nil, errors.New("error parsing timeWindow")
	}

	creditsPerTimeWindow := int(claims["rateLimiterCreditsPerTimeWindow"].(float64))

	return &Config{
		TimeWindow:           timeWindow,
		CreditsPerTimeWindow: creditsPerTimeWindow,
	}, nil
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
