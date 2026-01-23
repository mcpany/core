package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPRateLimitMiddleware_Security_Bypass(t *testing.T) {
	// 1 RPS, burst 1, Trust Proxy Enabled
	limiter := NewHTTPRateLimitMiddleware(1, 1, WithTrustProxy(true))
	handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	remoteAddr := "10.0.0.1:1234" // Proxy IP
	realClientIP := "203.0.113.1" // Attacker's real IP (as seen by proxy)

	// Attacker sends request 1
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = remoteAddr
	// Proxy appends realClientIP. Attacker sets "fake-ip-1"
	req1.Header.Set("X-Forwarded-For", "fake-ip-1, " + realClientIP)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code, "Request 1 should be allowed")

	// Attacker sends request 2 IMMEDIATELY
	// If rate limiting works on Real Client IP, this should be blocked (1 RPS)

	// Attacker changes fake IP to "fake-ip-2"
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = remoteAddr
	req2.Header.Set("X-Forwarded-For", "fake-ip-2, " + realClientIP)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusTooManyRequests, rec2.Code, "Request 2 should be blocked because it comes from the same real client IP")
}
