// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"time"

	"github.com/mcpany/core/pkg/util"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// HTTPRateLimitMiddleware provides global rate limiting for HTTP endpoints.
type HTTPRateLimitMiddleware struct {
	limiters *cache.Cache
	rps      rate.Limit
	burst    int
}

// NewHTTPRateLimitMiddleware creates a new HTTPRateLimitMiddleware.
// rps: requests per second per IP
// burst: maximum burst size per IP
func NewHTTPRateLimitMiddleware(rps float64, burst int) *HTTPRateLimitMiddleware {
	// Cleanup limiters every 10 minutes, expire after 5 minutes of inactivity
	return &HTTPRateLimitMiddleware{
		limiters: cache.New(5*time.Minute, 10*time.Minute),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// Handler wraps an http.Handler with rate limiting.
func (m *HTTPRateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := util.ExtractIP(r.RemoteAddr)

		var limiter *rate.Limiter
		if val, found := m.limiters.Get(ip); found {
			limiter = val.(*rate.Limiter)
		} else {
			limiter = rate.NewLimiter(m.rps, m.burst)
			m.limiters.Set(ip, limiter, cache.DefaultExpiration)
		}

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
