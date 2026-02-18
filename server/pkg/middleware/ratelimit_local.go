// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// LocalLimiter is an in-memory implementation of Limiter.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if a single request is allowed.
//
// Summary: Checks if one token is available.
//
// Parameters:
//   - _ : context.Context. Unused for local limiter (no network calls).
//
// Returns:
//   - bool: True if allowed, false otherwise.
//   - error: Always nil.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if a request with a specific cost is allowed.
//
// Summary: Checks if N tokens are available.
//
// Parameters:
//   - _ : context.Context. Unused.
//   - n: int. The number of tokens to consume.
//
// Returns:
//   - bool: True if allowed, false otherwise.
//   - error: Always nil.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update dynamically updates the rate limit configuration.
//
// Summary: Updates RPS and burst settings.
//
// Parameters:
//   - rps: float64. The new requests per second.
//   - burst: int. The new burst capacity.
func (l *LocalLimiter) Update(rps float64, burst int) {
	limit := rate.Limit(rps)
	if l.Limit() != limit {
		l.SetLimit(limit)
	}
	if l.Burst() != burst {
		l.SetBurst(burst)
	}
}

// LocalStrategy implements RateLimitStrategy for local in-memory rate limiting.
//
// Summary: Strategy for creating local rate limiters.
type LocalStrategy struct{}

// NewLocalStrategy creates a new LocalStrategy.
//
// Summary: Initializes a new LocalStrategy.
//
// Returns:
//   - *LocalStrategy: A new instance of LocalStrategy.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter based on the configuration.
//
// Summary: Creates a local rate limiter.
//
// Parameters:
//   - _ : context.Context. Unused.
//   - _ : string. Service ID (unused).
//   - _ : string. Limit scope key (unused).
//   - _ : string. Partition key (unused).
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - Limiter: The created local limiter.
//   - error: Always nil.
func (s *LocalStrategy) Create(_ context.Context, _, _, _ string, config *configv1.RateLimitConfig) (Limiter, error) {
	rps := config.GetRequestsPerSecond()
	burst := int(config.GetBurst())
	if burst <= 0 {
		burst = 1
	}
	return &LocalLimiter{
		Limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}, nil
}
