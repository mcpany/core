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

// HTTPRateLimitMiddleware represents the public HTTPRateLimitMiddleware entity.
//
// Summary: Defines the structured data model representing a rate limit middleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type HTTPRateLimitMiddleware struct {
	limiters   *ttlcache.Cache[string, *rate.Limiter]
	rps        rate.Limit
	burst      int
	trustProxy bool
}

// HTTPRateLimitOption represents the public HTTPRateLimitOption entity.
//
// Summary: Defines the structured data model representing a rate limit option.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type HTTPRateLimitOption func(*HTTPRateLimitMiddleware)

// WithTrustProxy serves as a public interface for interacting with WithTrustProxy.
//
// Summary: With the trust proxy appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func WithTrustProxy(trust bool) HTTPRateLimitOption {
	return func(m *HTTPRateLimitMiddleware) {
		m.trustProxy = trust
	}
}

// NewHTTPRateLimitMiddleware serves as a public interface for interacting with NewHTTPRateLimitMiddleware.
//
// Summary: Constructs and returns an initialized http rate limit middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Handler serves as a public interface for interacting with Handler.
//
// Summary: Handler the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
