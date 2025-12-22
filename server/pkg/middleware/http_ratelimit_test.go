// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHTTPRateLimitMiddleware_Burst(t *testing.T) {
	// 1 request per second, burst 1
	rl := middleware.NewHTTPRateLimitMiddleware(1, 1)
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(handler)
	defer ts.Close()
	client := ts.Client()

	req, _ := http.NewRequest("GET", ts.URL, nil)

	// 1. OK
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. Fail (Too Many Requests)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	// Wait 1.1s
	time.Sleep(1100 * time.Millisecond)

	// 3. OK
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
