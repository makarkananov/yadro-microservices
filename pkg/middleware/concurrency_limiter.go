package middleware

import "net/http"

// ConcurrencyLimiter is a middleware that limits the number of concurrent requests.
type ConcurrencyLimiter struct {
	semaphore chan struct{}
}

// NewConcurrencyLimiter creates a new instance of ConcurrencyLimiter.
func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		semaphore: make(chan struct{}, limit),
	}
}

// Limit limits the number of concurrent requests.
func (cl *ConcurrencyLimiter) Limit(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cl.semaphore <- struct{}{}
		defer func() {
			<-cl.semaphore
		}()
		handler.ServeHTTP(w, r)
	}
}
