package middleware

import (
	"go.uber.org/ratelimit"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a middleware that limits the number of requests per IP address.
// It uses the ratelimit package to implement the rate limiting logic.
type RateLimiter struct {
	rate    int
	per     time.Duration
	timeout time.Duration
	clients map[string]ratelimit.Limiter
	mu      sync.Mutex
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter(rate int, per time.Duration, timeout time.Duration) *RateLimiter {
	return &RateLimiter{
		rate:    rate,
		per:     per,
		timeout: timeout,
		clients: make(map[string]ratelimit.Limiter),
	}
}

// Limit limits the number of requests per IP address.
func (rl *RateLimiter) Limit(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rl.mu.Lock()
		if _, found := rl.clients[ip]; !found {
			rl.clients[ip] = ratelimit.New(rl.rate, ratelimit.Per(rl.per), ratelimit.WithoutSlack)
		}
		limiter := rl.clients[ip]
		rl.mu.Unlock()

		doneCh := make(chan struct{})

		go func() {
			defer close(doneCh)
			limiter.Take()
		}()

		select {
		case <-doneCh:
			handler.ServeHTTP(w, r)
			return
		case <-time.After(rl.timeout):
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
	}
}
