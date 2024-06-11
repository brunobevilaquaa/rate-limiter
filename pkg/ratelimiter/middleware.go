package ratelimiter

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

func (rh *UserRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheRepository := NewCacheRepository(rh.redisClient)

		config, err := rh.getConfigFromJWTToken(r)
		if err != nil {
			config = &rh.config
		}

		rl := NewLimiterService(cacheRepository, *config)
		key := rh.getUserKey(r)

		if allowed, err := rl.Allowed(key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if !allowed {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (rh *UserRateLimiter) getConfigFromJWTToken(r *http.Request) (*Config, error) {
	token := r.Header.Get("API_KEY")
	if token == "" {
		return nil, errors.New("missing API_KEY header")
	}

	data, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(rh.jwtSecret), nil
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

	timeWindow, err := time.ParseDuration(claims["timeWindow"].(string))
	if err != nil {
		return nil, errors.New("error parsing timeWindow")
	}

	creditsPerTimeWindow := int(claims["creditsPerTimeWindow"].(float64))

	return &Config{
		TimeWindow:           timeWindow,
		CreditsPerTimeWindow: creditsPerTimeWindow,
	}, nil
}

func getClientIP(r *http.Request) string {
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

func (rh *UserRateLimiter) getUserKey(r *http.Request) string {
	key := r.Header.Get("API_KEY")
	if key == "" {
		key = getClientIP(r)
	}

	hash := sha256.New()
	hash.Write([]byte(key))
	hashedKey := hash.Sum(nil)

	return hex.EncodeToString(hashedKey)
}
