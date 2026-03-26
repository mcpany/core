// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// LocalLimiter localLimiter represents a local limiter.
//
// Summary: LocalLimiter represents a local limiter.
type LocalLimiter struct {
	*rate.Limiter
}

// Allow checks if the request is allowed (cost 1).
//
// Summary: Checks if a single event is allowed by the rate limiter.
//
// Parameters: - None.
//   - _: context.Context. Unused.
//
// Returns: - None.
//   - bool: True if allowed, false otherwise.
//   - error: Always nil.
//
// Side Effects: - None.
//   - Consumes 1 token from the bucket if allowed.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Summary: Checks if N events are allowed by the rate limiter.
//
// Parameters: - None.
//   - _: context.Context. Unused.
//   - n: int. The cost of the event.
//
// Returns: - None.
//   - bool: True if allowed, false otherwise.
//   - error: Always nil.
//
// Side Effects: - None.
//   - Consumes n tokens from the bucket if allowed.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the .
//
// Summary: Updates the .
//
// Parameters:
//   - rps (float64): The rps.
//   - burst (int): The burst.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *LocalLimiter) Update(rps float64, burst int) {
	limit := rate.Limit(rps)
	if l.Limit() != limit {
		l.SetLimit(limit)
	}
	if l.Burst() != burst {
		l.SetBurst(burst)
	}
}

// LocalStrategy localStrategy represents a local strategy.
//
// Summary: LocalStrategy represents a local strategy.
type LocalStrategy struct{}

// NewLocalStrategy creates a new local strategy.
//
// Summary: Creates a new local strategy.
//
// Parameters:
//   - None.
//
// Returns:
//   - *LocalStrategy: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter.
//
// Summary: Creates a new in-memory rate limiter based on the provided configuration.
//
// Parameters: - None.
//   - _: context.Context. Unused.
//   - _: string. Unused (serviceID).
//   - _: string. Unused (limitScopeKey).
//   - _: string. Unused (partitionKey).
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns: - None.
//   - Limiter: The created LocalLimiter.
//   - error: Always nil.
//
// Side Effects: - None.
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
