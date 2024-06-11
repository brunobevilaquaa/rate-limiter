package ratelimiter

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value *UserDetail) error
	Get(ctx context.Context, key string) (*UserDetail, error)
}

type CacheRepository struct {
	cache *redis.Client
}

func NewCacheRepository(cache *redis.Client) *CacheRepository {
	return &CacheRepository{cache: cache}
}

func (cr *CacheRepository) Set(ctx context.Context, key string, value *UserDetail) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.New("error marshalling value")
	}

	if err := cr.cache.Set(ctx, key, data, 0).Err(); err != nil {
		return errors.New("error setting value in cache")
	}

	return nil
}

func (cr *CacheRepository) Get(ctx context.Context, key string) (*UserDetail, error) {
	val, err := cr.cache.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, errors.New("error getting value from cache")
	}

	var userDetail UserDetail
	if err = json.Unmarshal([]byte(val), &userDetail); err != nil {
		return nil, errors.New("error unmarshalling value")
	}

	return &userDetail, nil
}
