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

// Allow provides allow functionality.
//
// Summary: Allow.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (l *LocalLimiter) Allow(_ context.Context) (bool, error) {
	return l.Limiter.Allow(), nil
}

// AllowN provides allown functionality.
//
// Summary: AllowN.
//
// Parameters.
//   - _: The parameter.
//   - n: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (l *LocalLimiter) AllowN(_ context.Context, n int) (bool, error) {
	return l.Limiter.AllowN(time.Now(), n), nil
}

// Update provides update functionality.
//
// Summary: Update.
//
// Parameters.
//   - rps: The parameter.
//   - burst: The parameter.
//
// Returns.
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
// Summary: Strategy for creating local rate limiters.
type LocalStrategy struct{}

// NewLocalStrategy provides newlocalstrategy functionality.
//
// Summary: NewLocalStrategy.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewLocalStrategy() *LocalStrategy {
	return &LocalStrategy{}
}

// Create provides create functionality.
//
// Summary: Create.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - _: The parameter.
//   - _: The parameter.
//   - config: The parameter.
//   - error: The parameter.
//
// Returns.
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
