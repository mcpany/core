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
type Retry struct {
	config *configv1.RetryConfig
}

// NewRetry creates a new Retry instance with the given configuration.
// It sets default values for base and max backoff if they are not provided.
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

// Execute runs the provided work function, retrying it if it fails according
// to the configured policy.
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

	base := r.config.GetBaseBackoff().AsDuration()
	if base <= 0 {
		return 0
	}

	maxBackoff := r.config.GetMaxBackoff().AsDuration()

	var backoff time.Duration

	// Cap attempt to avoid potential overflow in 1<<attempt.
	// 62 is chosen because 1<<62 fits in int64 (positive).
	if attempt >= 62 {
		backoff = maxBackoff
	} else {
		// Calculate factor = 2^attempt
		factor := int64(1) << attempt

		// Check for overflow: base * factor > maxBackoff
		// factor > maxBackoff / base
		if factor > int64(maxBackoff/base) {
			backoff = maxBackoff
		} else {
			backoff = base * time.Duration(factor)
		}
	}

	// Add jitter (±20%)
	// Use float arithmetic for simplicity.
	// We randomize between 0.8 * backoff and 1.2 * backoff.
	// Note: rand.Float64() is global in math/rand for 1.20+ (seeded automatically in recent Go versions).
	jitter := time.Duration(float64(backoff) * (0.8 + 0.4*util.RandomFloat64()))
	return jitter
}
