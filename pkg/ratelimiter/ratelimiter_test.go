package ratelimiter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)
import "github.com/stretchr/testify/mock"

type CacheRepositoryMock struct {
	mock.Mock
}

func (m *CacheRepositoryMock) Set(ctx context.Context, key string, value *UserDetail) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *CacheRepositoryMock) Get(ctx context.Context, key string) (*UserDetail, error) {
	args := m.Called(ctx, key)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*UserDetail), args.Error(1)
}

func TestLimiterService_Allowed_ShouldAllowIfIsFirstReq(t *testing.T) {
	repository := &CacheRepositoryMock{}
	repository.On("Get", mock.Anything, mock.Anything).Return(nil, nil)
	repository.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)

	assert.NoError(t, err)
	assert.True(t, allowed)
	repository.AssertExpectations(t)
}

func TestLimiterService_Allowed_ShouldDecreaseCredits(t *testing.T) {
	repository := &CacheRepositoryMock{}
	firstRequestSentAt := time.Now().Add(-5 * time.Second)

	userDetail := &UserDetail{
		FirstRequestSentAt: firstRequestSentAt,
		RemainingCredits:   10,
	}

	repository.On("Get", mock.Anything, mock.Anything).Return(userDetail, nil)
	repository.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 9, userDetail.RemainingCredits)
}

func TestLimiterService_Allowed_ShouldBlockUserWhenLimitIsReached(t *testing.T) {
	repository := &CacheRepositoryMock{}
	firstRequestSentAt := time.Now().Add(-5 * time.Second)

	userDetail := &UserDetail{
		FirstRequestSentAt: firstRequestSentAt,
		RemainingCredits:   1,
	}

	repository.On("Get", mock.Anything, mock.Anything).Return(userDetail, nil)
	repository.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)
	assert.Equal(t, 0, userDetail.RemainingCredits)
	assert.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = limiter.Allowed(key)
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, userDetail.RemainingCredits)
	repository.AssertExpectations(t)
}

func TestLimiterService_Allowed_WhenUserReachedLimit(t *testing.T) {
	firstRequestSentAt := time.Now().Add(-5 * time.Second)
	userDetail := &UserDetail{
		FirstRequestSentAt: firstRequestSentAt,
		RemainingCredits:   0,
	}

	repository := &CacheRepositoryMock{}
	repository.On("Get", mock.Anything, mock.Anything).Return(userDetail, nil)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)

	assert.NoError(t, err)
	assert.False(t, allowed)
	repository.AssertExpectations(t)
}

func TestLimiterService_Allowed_ShouldAllowWhenTimeWindowIsRenewed(t *testing.T) {
	firstRequestSentAt := time.Now().Add(-11 * time.Second)
	userDetail := &UserDetail{
		FirstRequestSentAt: firstRequestSentAt,
		RemainingCredits:   0,
	}

	repository := &CacheRepositoryMock{}
	repository.On("Get", mock.Anything, mock.Anything).Return(userDetail, nil)
	repository.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 9, userDetail.RemainingCredits)
	repository.AssertExpectations(t)
}

func TestLimiterService_Allowed_ShouldReturnErrorWhenGetFails(t *testing.T) {
	repository := &CacheRepositoryMock{}
	repository.On("Get", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)

	assert.Error(t, err)
	assert.False(t, allowed)
	repository.AssertExpectations(t)
}

func TestLimiterService_Allowed_ShouldReturnErrorWhenSetFails(t *testing.T) {
	repository := &CacheRepositoryMock{}
	repository.On("Get", mock.Anything, mock.Anything).Return(nil, nil)
	repository.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	limiter := NewLimiterService(repository, Config{
		TimeWindow:           time.Duration(10) * time.Second,
		CreditsPerTimeWindow: 10,
	})

	key := "test"

	allowed, err := limiter.Allowed(key)

	assert.Error(t, err)
	assert.False(t, allowed)
	repository.AssertExpectations(t)
}
