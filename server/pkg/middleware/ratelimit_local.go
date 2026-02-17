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

// Allow checks if the request is allowed (cost 1).
//
// Summary: Checks if a request is allowed by the local rate limiter.
//
// Parameters:
//   - ctx: context.Context. The context (unused for local limiter).
//
// Returns:
//   - bool: True if the request is allowed, false otherwise.
//   - error: Always nil.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Summary: Checks if N requests are allowed by the local rate limiter.
//
// Parameters:
//   - ctx: context.Context. The context (unused).
//   - n: int. The number of tokens (cost) to consume.
//
// Returns:
//   - bool: True if the request is allowed, false otherwise.
//   - error: Always nil.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
//
// Summary: Updates the rate limit parameters.
//
// Parameters:
//   - rps: float64. The new requests per second limit.
//   - burst: int. The new burst limit.
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
type LocalStrategy struct{}

// NewLocalStrategy creates a new LocalStrategy.
//
// Returns the result.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter.
//
// Summary: Creates a new local rate limiter instance.
//
// Parameters:
//   - ctx: context.Context. The context (unused).
//   - _ : string. Unused.
//   - _ : string. Unused.
//   - _ : string. Unused.
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - Limiter: The created rate limiter.
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
