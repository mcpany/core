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
// Summary: Is an in-memory implementation of Limiter.
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

// Allow checks if the request is allowed (cost 1).
//
// Summary: Checks if the request is allowed (cost 1).
//
// Parameters:
//   - unnamed (context.Context): Parameter.
//
// Returns:
//   - bool: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN checks if the request is allowed with a specific cost.
//
// Summary: Checks if the request is allowed with a specific cost.
//
// Parameters:
//   - unnamed (context.Context): Parameter.
//   - n (int): Parameter.
//
// Returns:
//   - bool: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update updates the limiter configuration.
//
// Summary: Updates the limiter configuration.
//
// Parameters:
//   - rps (float64): Parameter.
//   - burst (int): Parameter.
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

// LocalStrategy implements RateLimitStrategy for local in-memory rate limiting.
//
// Summary: Implements RateLimitStrategy for local in-memory rate limiting.
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

// NewLocalStrategy creates a new LocalStrategy.
//
// Summary: Creates a new LocalStrategy.
//
// Parameters:
//   - None.
//
// Returns:
//   - *LocalStrategy: Return value.
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
// Summary: Creates a new LocalLimiter.
//
// Parameters:
//   - unnamed (context.Context): Parameter.
//   - unnamed (string): Parameter.
//   - unnamed (string): Parameter.
//   - unnamed (string): Parameter.
//   - config (*configv1.RateLimitConfig): Parameter.
//
// Returns:
//   - Limiter: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
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
