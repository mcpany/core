// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPRateLimitMiddleware_IPv6Normalization(t *testing.T) {
	// RPS = 1, Burst = 1.
	// We trust proxy to use XFF.
	mw := NewHTTPRateLimitMiddleware(1, 1, WithTrustProxy(true))
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request 1: [::1]
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("X-Forwarded-For", "[::1]")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("Request 1 failed: %d", rec1.Code)
	}

	// Request 2: ::1
	// Should be treated as SAME IP. Since Burst=1 and we just used 1, this should fail (or we wait).
	// With RPS=1, we need to wait 1 second to replenish. We don't wait.
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Forwarded-For", "::1")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	// If bug exists, [::1] and ::1 are different keys, so req2 succeeds.
	// If fix works, they are same key, so req2 fails (Too Many Requests).
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("Request 2 (normalized IP) passed with status %d, expected 429 (Too Many Requests). This implies [::1] and ::1 are treated as different IPs.", rec2.Code)
	}
}
