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
//
// Summary: In-memory rate limiter using token bucket.
//
// Fields:
//   - Limiter: *rate.Limiter. The underlying rate limiter.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed (cost 1).
//
// Summary: Checks if one token is available.
//
// Parameters:
//   - ctx: context.Context. Unused.
//
// Returns:
//   - bool: True if allowed.
//   - error: Always nil.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Summary: Checks if N tokens are available.
//
// Parameters:
//   - ctx: context.Context. Unused.
//   - n: int. The number of tokens.
//
// Returns:
//   - bool: True if allowed.
//   - error: Always nil.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
//
// Summary: Updates rate and burst.
//
// Parameters:
//   - rps: float64. New requests per second.
//   - burst: int. New burst size.
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
// Summary: Strategy for creating in-memory limiters.
type LocalStrategy struct{}

// NewLocalStrategy creates a new LocalStrategy.
//
// Summary: Initializes a LocalStrategy.
//
// Returns:
//   - *LocalStrategy: The initialized strategy.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter.
//
// Summary: Creates a new in-memory limiter based on config.
//
// Parameters:
//   - ctx: context.Context. Unused.
//   - serviceID: string. Unused.
//   - limitScopeKey: string. Unused.
//   - partitionKey: string. Unused.
//   - config: *configv1.RateLimitConfig. The configuration.
//
// Returns:
//   - Limiter: The created LocalLimiter.
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
