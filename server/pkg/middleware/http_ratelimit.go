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

// HTTPRateLimitMiddleware creates a new HTTP middleware that limits requests per IP.
// rps: requests per second allowed per IP.
// burst: maximum burst size allowed per IP.
func HTTPRateLimitMiddleware(rps float64, burst int) func(http.Handler) http.Handler {
	// Create a cache to hold limiters. Items expire after 1 hour of inactivity.
	// Cleanup runs every 10 minutes.
	limiters := cache.New(1*time.Hour, 10*time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := util.ExtractIP(r.RemoteAddr)

			// Get limiter for this IP
			var limiter *rate.Limiter
			if val, found := limiters.Get(ip); found {
				limiter = val.(*rate.Limiter)
			} else {
				newLimiter := rate.NewLimiter(rate.Limit(rps), burst)
				// Add only if it doesn't exist (handle race condition)
				if err := limiters.Add(ip, newLimiter, cache.DefaultExpiration); err == nil {
					limiter = newLimiter
				} else {
					// If Add failed, it means someone else added it concurrently.
					// Retrieve it again.
					if val, found := limiters.Get(ip); found {
						limiter = val.(*rate.Limiter)
					} else {
						// Should be extremely rare: added then expired/deleted immediately.
						// Fallback to new limiter (safe default, just resets bucket).
						limiter = newLimiter
					}
				}
			}

			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
