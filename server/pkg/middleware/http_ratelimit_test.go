// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPRateLimitMiddleware(t *testing.T) {
	// 5 RPS, Burst 1
	// We use a small burst to ensure we can hit the limit quickly.
	mw := HTTPRateLimitMiddleware(5.0, 1)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := ts.Client()

	// 1st request should succeed (Consumes the 1 token in burst)
	resp, err := client.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 2nd request should fail immediately (Burst empty, refill rate 5/s means 200ms per token)
	resp, err = client.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	// Wait for refill (250ms > 200ms)
	time.Sleep(250 * time.Millisecond)

	// 3rd request should succeed
	resp, err = client.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPRateLimitMiddleware_DifferentIPs(t *testing.T) {
	// 1 RPS, Burst 1
	mw := HTTPRateLimitMiddleware(1.0, 1)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// We can't easily spoof RemoteAddr with httptest.NewServer client.
	// We'll use direct ServeHTTP calls.

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/", nil)
	r1.RemoteAddr = "192.168.1.1:1234"

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.RemoteAddr = "192.168.1.1:1235" // Same IP, different port

	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/", nil)
	r3.RemoteAddr = "192.168.1.2:1234" // Different IP

	// Request 1 from IP 1
	handler.ServeHTTP(w1, r1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Request 2 from IP 1 (should fail)
	handler.ServeHTTP(w2, r2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Request 1 from IP 2 (should succeed)
	handler.ServeHTTP(w3, r3)
	assert.Equal(t, http.StatusOK, w3.Code)
}
