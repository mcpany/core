// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Summary: LocalLimiter is an in-memory implementation of Limiter. Rate limiter implementation using golang.org/x/time/rate.
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
type LocalLimiter struct {
	*rate.Limiter
}

// Summary: Allow checks if the request is allowed (cost 1). Checks if a single event is allowed by the rate limiter.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - bool: The resulting bool.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// Summary: AllowN checks if the request is allowed with a specific cost. Checks if N events are allowed by the rate limiter.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - n (int): The n parameter.
//
// Returns:
//   - bool: The resulting bool.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Summary: Update updates the limiter configuration. Dynamically updates the rate limit and burst size.
//
// Parameters:
//   - rps (float64): The rps parameter.
//   - burst (int): The burst parameter.
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

// Summary: LocalStrategy implements RateLimitStrategy for local in-memory rate limiting. Strategy for creating local rate limiters.
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
type LocalStrategy struct{}

// Summary: NewLocalStrategy creates a new LocalStrategy. Initializes a new LocalStrategy.
//
// Parameters:
//   - None.
//
// Returns:
//   - *LocalStrategy: The resulting *LocalStrategy.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Summary: Create creates a new LocalLimiter. Creates a new in-memory rate limiter based on the provided configuration.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (string): The _ parameter.
//   - _ (string): The _ parameter.
//   - _ (string): The _ parameter.
//   - config (*configv1.RateLimitConfig): The config parameter.
//
// Returns:
//   - Limiter: The resulting Limiter.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
