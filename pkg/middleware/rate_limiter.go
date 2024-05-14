package middleware

import (
	"net"
	"net/http"
	"sync"
	"yadro-microservices/pkg/limiter"
)

// RateLimiter is a middleware that limits the number of requests per IP address.
// It uses the /x/time/rate package to implement the rate limiting logic.
type RateLimiter struct {
	rate      int64
	maxTokens int64
	clients   map[string]*limiter.TokenBucket
	mu        sync.Mutex
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter(rate, maxTokens int64) *RateLimiter {
	return &RateLimiter{
		rate:      rate,
		maxTokens: maxTokens,
		clients:   make(map[string]*limiter.TokenBucket),
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
			rl.clients[ip] = limiter.NewTokenBucket(rl.rate, rl.maxTokens)
		}
		l := rl.clients[ip]
		rl.mu.Unlock()

		if l.Allow() {
			handler.ServeHTTP(w, r)
		} else {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}
	}
}
