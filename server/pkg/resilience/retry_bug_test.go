// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestBackoffJitterAtCap demonstrates that currently jitter is NOT applied when maxBackoff is reached.
func TestBackoffJitterAtCap(t *testing.T) {
	config := &configv1.RetryConfig{}
	config.SetBaseBackoff(durationpb.New(1 * time.Second))
	config.SetMaxBackoff(durationpb.New(10 * time.Second))
	retry := NewRetry(config)

	// Base = 1s.
	// Attempt 0: 1s.
	// Attempt 1: 2s.
	// Attempt 2: 4s.
	// Attempt 3: 8s.
	// Attempt 4: 16s -> Capped at 10s.

	// Check attempt 4. It should be capped.
	// Current behavior: It returns exactly 10s (no jitter).
	val := retry.backoff(4)

	// New expectation: jittered around 10s.
	// We assert it is NOT exactly 10s (assuming random jitter > 0).
	// Note: It's technically possible to hit exactly 1.0 jitter, but very unlikely with float64.
	// With the fix, this assertion ensures we are applying jitter.
	assert.NotEqual(t, 10*time.Second, val, "Expected jitter to be applied at cap")
	assert.InDelta(t, float64(10*time.Second), float64(val), float64(2500*time.Millisecond))
}

func TestRetry_NegativeRetries(t *testing.T) {
	ctx := context.Background()
	var attempts int
	work := func(_ context.Context) error {
		attempts++
		return nil // Success immediately
	}

	// -1 retries should be treated as 0 retries (1 attempt total)
	config := &configv1.RetryConfig{}
	config.SetNumberOfRetries(-1)
	retry := NewRetry(config)
	err := retry.Execute(ctx, work)
	require.NoError(t, err)
	require.Equal(t, 1, attempts)
}

func TestRetry_ContextDoneDuringBackoff(t *testing.T) {
	// We want to trigger the case <-time.After branch, but then have context cancelled?
	// No, we want to hit case <-ctx.Done().

	// If we set a long backoff, and cancel context in background.
	ctx, cancel := context.WithCancel(context.Background())

	var attempts int
	work := func(_ context.Context) error {
		attempts++
		// Fail first time
		if attempts == 1 {
			// Trigger cancel immediately after first failure (before sleep)
			// Wait, if we cancel here, the check `if ctx.Err() != nil` at start of loop
			// might catch it on next iteration, OR the select catches it.
			// The loop is: check ctx -> work -> check error -> check last attempt -> select.

			// If we cancel here, we are inside work.
			// Next step: return error.
			// Then check last attempt (not last).
			// Then select.
			// Since we cancelled, select should pick ctx.Done() immediately (or very fast).
			cancel()
			return assert.AnError
		}
		return nil
	}

	config := &configv1.RetryConfig{}
	config.SetNumberOfRetries(5)
	config.SetBaseBackoff(durationpb.New(1 * time.Hour)) // Long backoff
	retry := NewRetry(config)

	err := retry.Execute(ctx, work)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
	require.Equal(t, 1, attempts)
}
