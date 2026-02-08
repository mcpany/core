// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// HTTPRateLimitMiddleware provides global rate limiting for HTTP endpoints.
//
// Summary: provides global rate limiting for HTTP endpoints.
type HTTPRateLimitMiddleware struct {
	limiters   *cache.Cache
	rps        rate.Limit
	burst      int
	trustProxy bool
}

// HTTPRateLimitOption defines a functional option for HTTPRateLimitMiddleware.
//
// Summary: defines a functional option for HTTPRateLimitMiddleware.
type HTTPRateLimitOption func(*HTTPRateLimitMiddleware)

// WithTrustProxy enables trusting the X-Forwarded-For header.
//
// Summary: enables trusting the X-Forwarded-For header.
//
// Parameters:
//   - trust: bool. The trust.
//
// Returns:
//   - HTTPRateLimitOption: The HTTPRateLimitOption.
func WithTrustProxy(trust bool) HTTPRateLimitOption {
	return func(m *HTTPRateLimitMiddleware) {
		m.trustProxy = trust
	}
}

// NewHTTPRateLimitMiddleware creates a new HTTPRateLimitMiddleware.
//
// Summary: creates a new HTTPRateLimitMiddleware.
//
// Parameters:
//   - rps: float64. The rps.
//   - burst: int. The burst.
//   - opts: ...HTTPRateLimitOption. The opts.
//
// Returns:
//   - *HTTPRateLimitMiddleware: The *HTTPRateLimitMiddleware.
func NewHTTPRateLimitMiddleware(rps float64, burst int, opts ...HTTPRateLimitOption) *HTTPRateLimitMiddleware {
	// Cleanup limiters every 10 minutes, expire after 5 minutes of inactivity
	m := &HTTPRateLimitMiddleware{
		limiters: cache.New(5*time.Minute, 10*time.Minute),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Handler wraps an http.Handler with rate limiting.
//
// Summary: wraps an http.Handler with rate limiting.
//
// Parameters:
//   - next: http.Handler. The next.
//
// Returns:
//   - http.Handler: The http.Handler.
func (m *HTTPRateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := util.ExtractIP(r.RemoteAddr)

		if m.trustProxy {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// Use the last IP in the list (the IP that connected to the trusted proxy).
				// Standard proxies append the connecting IP to the list.
				// We trust the proxy to have appended the correct IP, but we do NOT trust the
				// earlier IPs in the list as they can be spoofed by the client.
				if idx := strings.LastIndex(xff, ","); idx != -1 {
					ip = strings.TrimSpace(xff[idx+1:])
				} else {
					ip = strings.TrimSpace(xff)
				}
			}
		}

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
