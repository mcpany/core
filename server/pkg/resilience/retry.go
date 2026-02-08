// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"errors"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Retry implements a retry policy for failed operations.
//
// Summary: implements a retry policy for failed operations.
type Retry struct {
	config *configv1.RetryConfig
}

// NewRetry creates a new Retry instance with the given configuration.
//
// Summary: creates a new Retry instance with the given configuration.
//
// Parameters:
//   - config: *configv1.RetryConfig. The config.
//
// Returns:
//   - *Retry: The *Retry.
func NewRetry(config *configv1.RetryConfig) *Retry {
	if config == nil {
		config = &configv1.RetryConfig{}
	}
	if config.GetBaseBackoff() == nil {
		config.SetBaseBackoff(durationpb.New(time.Second))
	}
	if config.GetMaxBackoff() == nil {
		config.SetMaxBackoff(durationpb.New(30 * time.Second))
	}
	return &Retry{
		config: config,
	}
}

// Execute runs the provided work function, retrying it if it fails according.
//
// Summary: runs the provided work function, retrying it if it fails according.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - work: func(context.Context) error. The work.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (r *Retry) Execute(ctx context.Context, work func(context.Context) error) error {
	var err error
	// Use int64 for attempts to match usage, though retries count is usually small.
	// Cast safely.
	retries := int(r.config.GetNumberOfRetries())
	if retries < 0 {
		retries = 0
	}

	for i := 0; i < retries+1; i++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = work(ctx)
		if err == nil {
			return nil
		}

		var permanentErr *PermanentError
		if errors.As(err, &permanentErr) {
			return err
		}

		// If this was the last attempt, return the error immediately
		if i == retries {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.backoff(i)):
			// continue
		}
	}
	return err
}

func (r *Retry) backoff(attempt int) time.Duration {
	if attempt < 0 {
		return r.config.GetBaseBackoff().AsDuration()
	}

	// Cap attempt to avoid potential overflow in 1<<attempt.
	// 62 is chosen because 1<<62 fits in int64 (positive).
	// With base=1ns, 1<<62 ns is > 100 years, far exceeding any reasonable MaxBackoff.
	if attempt >= 62 {
		return r.config.GetMaxBackoff().AsDuration()
	}

	base := r.config.GetBaseBackoff().AsDuration()
	maxBackoff := r.config.GetMaxBackoff().AsDuration()

	if base <= 0 {
		return 0
	}

	// Calculate factor = 2^attempt
	factor := int64(1) << attempt

	// Check for overflow: base * factor > maxBackoff
	// factor > maxBackoff / base
	if factor > int64(maxBackoff/base) {
		return maxBackoff
	}

	backoff := base * time.Duration(factor)

	// Add jitter (±20%)
	// Use float arithmetic for simplicity, or integer math to avoid issues.
	// We'll use a simple approach: randomize between 0.8 * backoff and 1.2 * backoff.
	// Or just "Full Jitter": random(0, backoff).
	// But commonly we want to stay close to exponential.
	// Let's do ±20% jitter.
	// Note: rand.Float64() is global in math/rand for 1.20+ (seeded automatically in recent Go versions? No, need Seed).
	// But math/rand/v2 is 1.22.
	// We'll assume we can use math/rand with shared source or just time.
	// better: backoff = backoff + random(0, backoff/2) to always increase?
	// or just +/-.

	// Implementation:
	// We avoid global rand issues by using a local source or crypto/rand if needed, but for backoff math/rand is fine.
	// However, we don't want to init rand source every time.
	// We'll trust global rand is seeded or acceptable for jitter.
	// Actually, relying on global rand without Seed might be deterministic.
	// Go 1.20+ seeds global rand automatically.

	jitter := time.Duration(float64(backoff) * (0.8 + 0.4*util.RandomFloat64()))
	return jitter
}
