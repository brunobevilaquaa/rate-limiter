# Rate Limiter

This project implements a rate limiter in Go that can be configured to limit the maximum number of requests per second based on a specific IP address or an access token.

## Features

- Request limiting by IP address.
- Request limiting by access token (configurable via the header `API_KEY: <TOKEN>`).
- Configuration of limits.
- Middleware for web server integration.
- Appropriate responses when the limit is exceeded (HTTP 429 status code and specific message).
- Storage of limiting information in a Redis database (using Docker Compose).
- Structure for easy swapping of Redis for another persistence mechanism.

### Example

```go
rdb := redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "",
	DB:       0,
})

cache := ratelimiter.NewCacheRepository(rdb)
creditsPerToken := 10
creditsPerIp := 5
timeWindow := time.Second * 10

middleware := ratelimiter.NewUserRateLimiter(mux, cache, creditsPerToken, creditsPerIp, timeWindow)
```

For more examples, see the [examples directory](examples/ratelimiter/).

## Requirements

- Go 1.22+
- Redis

## Configuration

## Default Configuration

Rate limiter can be configured with the following parameters:

```go
ratelimiter.Config{
    TimeWindow:           time.Second,
    CreditsPerTimeWindow: 10,
}
```

## Middleware configuration

The middleware can be configured with the following parameters:

```go
ratelimiter.MiddlewareConfig{
    TimeWindow:      time.Second * 10,
    CreditsPerIp:    5,
    CreditsPerToken: 10,
}

```

### Persistence Structure

The rate limiter logic is separated from the middleware, allowing easy swapping of the persistence mechanism. The default implementation uses Redis, but it can be replaced with another mechanism as needed.

The cache repository can be implemented by creating a new struct that satisfies the `CacheRepository` interface:

```go
type Cache interface {
    Set(ctx context.Context, key string, value *ratelimiter.UserDetail) error
    Get(ctx context.Context, key string) (*ratelimiter.UserDetail, error)
}
```

## Error Responses

When the limit is exceeded, the system responds with:

- HTTP Status Code: 429
- Message: "you have reached the maximum number of requests or actions allowed within a certain time frame"

## Use Cases

### Limiting by IP

Suppose the rate limiter is configured to allow a maximum of 5 requests per second per IP. If the IP `192.168.1.1` sends 6 requests in one second, the sixth request should be blocked.

### Limiting by Token

If a token has a configured limit of 10 requests per second and sends 11 requests in that interval, the eleventh request should be blocked.

In both cases, subsequent requests can be made only after the block duration has expired.