package ratelimiter

import (
	"context"
	"time"
)

type Config struct {
	TimeWindow           time.Duration
	CreditsPerTimeWindow int
}

type UserDetail struct {
	FirstRequestSentAt time.Time
	RemainingCredits   int
}

type LimiterService struct {
	cacheRepository Cache
	config          Config
}

func NewLimiterService(cacheRepository Cache, config Config) *LimiterService {
	return &LimiterService{
		cacheRepository: cacheRepository,
		config:          config,
	}
}

func (ls *LimiterService) Allowed(key string) (bool, error) {
	ctx := context.Background()

	userDetail, err := ls.cacheRepository.Get(ctx, key)
	if err != nil {
		return false, err
	}

	if userDetail == nil || userDetail.FirstRequestSentAt.IsZero() {
		userDetail = &UserDetail{
			FirstRequestSentAt: time.Now(),
			RemainingCredits:   ls.config.CreditsPerTimeWindow,
		}

		if err := ls.cacheRepository.Set(ctx, key, userDetail); err != nil {
			return false, err
		}

		return true, nil
	}

	timeSinceFirstRequest := time.Now().Sub(userDetail.FirstRequestSentAt)
	timeWindow := ls.config.TimeWindow

	if timeSinceFirstRequest < timeWindow {
		if userDetail.RemainingCredits <= 0 {
			return false, nil
		}

		userDetail.RemainingCredits--
	} else {
		userDetail.FirstRequestSentAt = time.Now()
		userDetail.RemainingCredits = ls.config.CreditsPerTimeWindow
	}

	if err := ls.cacheRepository.Set(ctx, key, userDetail); err != nil {
		return false, err
	}

	return true, nil
}
