package middleware

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestChain(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		assert.NoError(t, err)
	}

	mw1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "1")
			next.ServeHTTP(w, r)
		}
	}

	mw2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "2")
			next.ServeHTTP(w, r)
		}
	}

	chain := Chain(handler, mw1, mw2)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	chain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "1", rr.Header().Get("X-Middleware-1"))
	assert.Equal(t, "2", rr.Header().Get("X-Middleware-2"))
	assert.Equal(t, "OK", rr.Body.String())
}

func TestRateLimiter(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		assert.NoError(t, err)
	}

	limiter := NewRateLimiter(1, 1)

	limitedHandler := limiter.Limit(http.HandlerFunc(handler))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	rr1 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr1, req)

	rr2 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr2, req)

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, "OK", rr1.Body.String())

	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
	assert.Contains(t, rr2.Body.String(), "Too many requests")

	time.Sleep(1 * time.Second)

	rr3 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr3, req)

	assert.Equal(t, http.StatusOK, rr3.Code)
	assert.Equal(t, "OK", rr3.Body.String())
}

func TestRateLimiterMultipleIPs(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		assert.NoError(t, err)
	}

	limiter := NewRateLimiter(1, 1)

	limitedHandler := limiter.Limit(http.HandlerFunc(handler))

	req1, _ := http.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:1234"

	req2, _ := http.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.2:1234"

	rr1 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr1, req1)

	rr2 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr2, req2)

	rr3 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr3, req1)

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, "OK", rr1.Body.String())

	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, "OK", rr2.Body.String())

	assert.Equal(t, http.StatusTooManyRequests, rr3.Code)
	assert.Contains(t, rr3.Body.String(), "Too many requests")
}

func TestConcurrencyLimiter(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		assert.NoError(t, err)
	}

	limiter := NewConcurrencyLimiter(1)

	limitedHandler := limiter.Limit(http.HandlerFunc(handler))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	var wg sync.WaitGroup
	wg.Add(2)

	var rr1, rr2 *httptest.ResponseRecorder

	go func() {
		defer wg.Done()
		rr1 = httptest.NewRecorder()
		limitedHandler.ServeHTTP(rr1, req)
	}()

	time.Sleep(100 * time.Millisecond)

	go func() {
		defer wg.Done()
		rr2 = httptest.NewRecorder()
		limitedHandler.ServeHTTP(rr2, req)
	}()

	wg.Wait()

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, "OK", rr1.Body.String())

	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, "OK", rr2.Body.String())
}
