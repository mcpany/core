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

// HTTPRateLimitMiddleware limits the number of requests per client IP.
type HTTPRateLimitMiddleware struct {
	limiters *cache.Cache
	rps      rate.Limit
	burst    int
}

// NewHTTPRateLimitMiddleware creates a new middleware with given RPS and burst.
func NewHTTPRateLimitMiddleware(rps float64, burst int) *HTTPRateLimitMiddleware {
	return &HTTPRateLimitMiddleware{
		limiters: cache.New(10*time.Minute, 20*time.Minute),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// Handler returns a handler that enforces rate limiting.
func (m *HTTPRateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := util.ExtractIP(r.RemoteAddr)

		// Get or create limiter
		var limiter *rate.Limiter
		if val, found := m.limiters.Get(ip); found {
			limiter = val.(*rate.Limiter)
		} else {
			limiter = rate.NewLimiter(m.rps, m.burst)
			m.limiters.Set(ip, limiter, cache.DefaultExpiration)
		}

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
