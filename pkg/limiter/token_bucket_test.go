package limiter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(1, 10)

	for i := 0; i < 10; i++ {
		assert.True(t, tb.Allow(), "Initial tokens should allow requests")
	}

	assert.False(t, tb.Allow(), "Should not allow more requests than initial tokens")

	time.Sleep(1 * time.Second)
	assert.True(t, tb.Allow(), "Should allow one request after 1 second")

	time.Sleep(1 * time.Second)
	assert.True(t, tb.Allow(), "Should allow another request after another second")

	time.Sleep(10 * time.Second)
	for i := 0; i < 10; i++ {
		assert.True(t, tb.Allow(), "Should allow up to max tokens requests")
	}
	assert.False(t, tb.Allow(), "Should not allow more requests than max tokens")
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(2, 5)

	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow(), "Initial tokens should allow requests")
	}

	assert.False(t, tb.Allow(), "Should not allow more requests than initial tokens")

	time.Sleep(2500 * time.Millisecond)
	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow(), "Should allow up to max tokens requests after refill")
	}
	assert.False(t, tb.Allow(), "Should not allow more requests than max tokens after refill")
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(1, 10)

	var allowed int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			if tb.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, 10, allowed, "Concurrent access should allow up to max tokens requests")
}
