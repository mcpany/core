// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestHTTPRateLimitMiddleware_LoopbackOrigin(t *testing.T) {
	// 2 RPS, burst 2
	limiter := NewHTTPRateLimitMiddleware(2, 2)
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First Origin
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Origin 1 Request %d should be allowed", i)
	}

	// 3rd request from First Origin should be blocked
	reqBlocked := httptest.NewRequest("GET", "/", nil)
	reqBlocked.RemoteAddr = "127.0.0.1:1234"
	reqBlocked.Header.Set("Origin", "http://example.com")
	recBlocked := httptest.NewRecorder()
	handler.ServeHTTP(recBlocked, reqBlocked)
	assert.Equal(t, http.StatusTooManyRequests, recBlocked.Code, "Origin 1 Request 3 should be blocked")

	// Second Origin should still be allowed, because loopback is rate-limited per origin
	reqDifferentOrigin := httptest.NewRequest("GET", "/", nil)
	reqDifferentOrigin.RemoteAddr = "127.0.0.1:1234"
	reqDifferentOrigin.Header.Set("Origin", "http://other.com")
	recDifferentOrigin := httptest.NewRecorder()
	handler.ServeHTTP(recDifferentOrigin, reqDifferentOrigin)
	assert.Equal(t, http.StatusOK, recDifferentOrigin.Code, "Origin 2 Request 1 should be allowed")

	// IPv6 Loopback
	reqIPv6 := httptest.NewRequest("GET", "/", nil)
	reqIPv6.RemoteAddr = "[::1]:1234"
	reqIPv6.Header.Set("Origin", "http://example.com")
	recIPv6 := httptest.NewRecorder()
	handler.ServeHTTP(recIPv6, reqIPv6)
	// Blocked because `::1` origin example.com shares the loopback key cache? Actually it's just loopback:http://example.com
	assert.Equal(t, http.StatusTooManyRequests, recIPv6.Code, "IPv6 Loopback Origin 1 should share the cache and be blocked")
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
