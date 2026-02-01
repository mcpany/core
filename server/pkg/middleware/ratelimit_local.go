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
// Parameters:
//   - _: The context (unused).
//
// Returns:
//   - bool: True if the request is allowed.
//   - error: An error if the operation fails (always nil for LocalLimiter).
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Parameters:
//   - _: The context (unused).
//   - n: The cost of the request.
//
// Returns:
//   - bool: True if the request is allowed.
//   - error: An error if the operation fails (always nil for LocalLimiter).
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
//
// Parameters:
//   - rps: The new requests per second.
//   - burst: The new burst size.
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
// Returns:
//   - *LocalStrategy: A new LocalStrategy instance.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create creates a new LocalLimiter.
//
// Parameters:
//   - _: The context (unused).
//   - _: The service ID (unused).
//   - _: The limit scope key (unused).
//   - _: The partition key (unused).
//   - config: The rate limit configuration.
//
// Returns:
//   - Limiter: A new LocalLimiter.
//   - error: An error if the operation fails.
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
