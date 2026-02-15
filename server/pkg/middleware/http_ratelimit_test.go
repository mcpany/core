package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPRateLimitMiddleware(t *testing.T) {
	// 5 RPS, burst 5
	limiter := NewHTTPRateLimitMiddleware(5, 5)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Allow first 5 requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should be allowed", i)
	}

	// 6th request should be blocked (burst exceeded)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "Request 6 should be blocked")

	// Different IP should be allowed
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code, "Different IP should be allowed")
}

func TestHTTPRateLimitMiddleware_TrustProxy(t *testing.T) {
	// 5 RPS, burst 5, Trust Proxy Enabled
	limiter := NewHTTPRateLimitMiddleware(5, 5, WithTrustProxy(true))
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Simulate requests from a load balancer (same RemoteAddr) but different X-Forwarded-For
	remoteAddr := "10.0.0.1:1234" // Proxy IP

	// User 1: 5 requests allowed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = remoteAddr
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "User 1 Request %d should be allowed", i)
	}

	// User 1: 6th request blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = remoteAddr
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "User 1 Request 6 should be blocked")

	// User 2: Should be allowed despite same RemoteAddr, because X-Forwarded-For is different
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = remoteAddr
	req2.Header.Set("X-Forwarded-For", "203.0.113.2")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code, "User 2 Request 1 should be allowed")

	// User 3: Multiple IPs in X-Forwarded-For
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.RemoteAddr = remoteAddr
	req3.Header.Set("X-Forwarded-For", "203.0.113.3, 10.0.0.1")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code, "User 3 Request 1 should be allowed")
}

func TestHTTPRateLimitMiddleware_NoTrustProxy(t *testing.T) {
	// 5 RPS, burst 5, Trust Proxy Disabled (Default)
	limiter := NewHTTPRateLimitMiddleware(5, 5, WithTrustProxy(false))
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Simulate requests from a load balancer (same RemoteAddr) but different X-Forwarded-For
	remoteAddr := "10.0.0.1:1234" // Proxy IP

	// User 1: 5 requests allowed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = remoteAddr
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "User 1 Request %d should be allowed", i)
	}

	// User 1: 6th request blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = remoteAddr
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "User 1 Request 6 should be blocked")

	// User 2: Should be BLOCKED because trustProxy is false, so it sees the same RemoteAddr
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = remoteAddr
	req2.Header.Set("X-Forwarded-For", "203.0.113.2")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	// This confirms that without TrustProxy, rate limiting is shared (Vulnerable behavior if behind proxy)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code, "User 2 Request 1 should be blocked due to shared IP")
}
