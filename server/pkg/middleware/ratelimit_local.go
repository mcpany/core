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
// Summary: Rate limiter implementation using golang.org/x/time/rate.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed (cost 1).
//
// Summary: Checks if a single event is allowed by the rate limiter.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - bool: True if allowed, false otherwise.
//   - error: Always nil.
//
// Side Effects:
//   - Consumes 1 token from the bucket if allowed.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Summary: Checks if N events are allowed by the rate limiter.
//
// Parameters:
//   - _: context.Context. Unused.
//   - n: int. The cost of the event.
//
// Returns:
//   - bool: True if allowed, false otherwise.
//   - error: Always nil.
//
// Side Effects:
//   - Consumes n tokens from the bucket if allowed.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
//
// Summary: Dynamically updates the rate limit and burst size.
//
// Parameters:
//   - rps: float64. The new requests per second limit.
//   - burst: int. The new burst size.
//
// Side Effects:
//   - Modifies the underlying rate.Limiter state.
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
//   - *LocalStrategy: The initialized strategy.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter.
//
// Summary: Creates a new in-memory rate limiter based on the provided configuration.
//
// Parameters:
//   - _: context.Context. Unused.
//   - _: string. Unused (serviceID).
//   - _: string. Unused (limitScopeKey).
//   - _: string. Unused (partitionKey).
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - Limiter: The created LocalLimiter.
//   - error: Always nil.
//
// Side Effects:
//   - Sets a minimum burst of 1 if configured lower.
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
