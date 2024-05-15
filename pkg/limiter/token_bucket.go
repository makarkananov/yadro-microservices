package limiter

import (
	"sync"
	"time"
)

// TokenBucket is a token bucket algorithm implementation.
type TokenBucket struct {
	rate       int64
	maxTokens  int64
	nowTokens  int64
	lastRefill time.Time
	mux        sync.Mutex
}

// NewTokenBucket creates a new TokenBucket instance.
func NewTokenBucket(rate int64, maxTokens int64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		maxTokens:  maxTokens,
		nowTokens:  maxTokens,
		lastRefill: time.Now(),
	}
}

// Allow checks if the token bucket has enough tokens to allow a request.
func (tb *TokenBucket) Allow() bool {
	tb.mux.Lock()
	defer tb.mux.Unlock()

	tb.refill()
	if tb.nowTokens > 0 {
		tb.nowTokens--
		return true
	}

	return false
}

// refill refills the token bucket with new tokens.
func (tb *TokenBucket) refill() {
	now := time.Now()
	end := time.Since(tb.lastRefill)
	needTokens := (end.Nanoseconds() * tb.rate) / int64(time.Second)
	tb.nowTokens = min(tb.maxTokens, tb.nowTokens+needTokens)
	tb.lastRefill = now
}
