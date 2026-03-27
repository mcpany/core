// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"golang.org/x/time/rate"
)

// HTTPRateLimitMiddleware provides global rate limiting for HTTP endpoints.
//
// Summary. Middleware for rate limiting HTTP requests based on IP address.
type HTTPRateLimitMiddleware struct {
	limiters   *ttlcache.Cache[string, *rate.Limiter]
	rps        rate.Limit
	burst      int
	trustProxy bool
}

// HTTPRateLimitOption defines a functional option for HTTPRateLimitMiddleware.
//
// Summary: Functional option type for configuring the middleware.
type HTTPRateLimitOption func(*HTTPRateLimitMiddleware)

// WithTrustProxy provides withtrustproxy functionality.
//
// Summary: WithTrustProxy.
//
// Parameters.
//   - trust: The parameter.
//
// Returns.
//   - result: The result.
func WithTrustProxy(trust bool) HTTPRateLimitOption {
	return func(m *HTTPRateLimitMiddleware) {
		m.trustProxy = trust
	}
}

// NewHTTPRateLimitMiddleware provides newhttpratelimitmiddleware functionality.
//
// Summary: NewHTTPRateLimitMiddleware.
//
// Parameters.
//   - rps: The parameter.
//   - burst: The parameter.
//   - opts: The parameter.
//
// Returns.
//   - result: The result.
func NewHTTPRateLimitMiddleware(rps float64, burst int, opts ...HTTPRateLimitOption) *HTTPRateLimitMiddleware {
	// ⚡ BOLT: Prevented unbounded memory growth by enforcing a capacity limit on the rate limiter cache.
	// Randomized Selection from Top 5 High-Impact Targets
	limiters := ttlcache.New[string, *rate.Limiter](
		ttlcache.WithTTL[string, *rate.Limiter](5*time.Minute),
		ttlcache.WithCapacity[string, *rate.Limiter](100000),
	)

	// Start the cache cleaner in a goroutine
	go limiters.Start()

	m := &HTTPRateLimitMiddleware{
		limiters: limiters,
		rps:      rate.Limit(rps),
		burst:    burst,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Handler provides handler functionality.
//
// Summary: Handler.
//
// Parameters.
//   - next: The parameter.
//
// Returns.
//   - result: The result.
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

		// Local-Loopback Rate Limiting check
		// For loopback IPs, we limit per-origin to prevent malicious websites from brute-forcing
		// local agent listeners. This fulfills the "Local Zero Trust" mandate.
		cacheKey := ip
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil && parsedIP.IsLoopback() {
			origin := r.Header.Get("Origin")
			if origin == "" {
				cacheKey = "loopback:no-origin"
			} else {
				cacheKey = "loopback:" + origin
			}
		}

		var limiter *rate.Limiter
		if item := m.limiters.Get(cacheKey); item != nil {
			limiter = item.Value()
		} else {
			limiter = rate.NewLimiter(m.rps, m.burst)
			m.limiters.Set(cacheKey, limiter, ttlcache.DefaultTTL)
		}

		if !limiter.Allow() {
			if strings.HasPrefix(cacheKey, "loopback:") {
				logging.GetLogger().Warn("Loopback rate limit exceeded", "origin", r.Header.Get("Origin"), "ip", ip)
			}
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
