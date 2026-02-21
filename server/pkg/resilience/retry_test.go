// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRetry(t *testing.T) {
	ctx := context.Background()
	t.Run("success_on_first_try", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			return nil
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		retry := NewRetry(config)
		err := retry.Execute(ctx, work)
		require.NoError(t, err)
		require.Equal(t, 1, attempts)
	})

	t.Run("success_after_retries", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("transient error")
			}
			return nil
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		config.SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		retry := NewRetry(config)
		err := retry.Execute(ctx, work)
		require.NoError(t, err)
		require.Equal(t, 3, attempts)
	})

	t.Run("failure_after_all_retries", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			return errors.New("persistent error")
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		config.SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		retry := NewRetry(config)
		err := retry.Execute(ctx, work)
		require.Error(t, err)
		require.Equal(t, 4, attempts)
	})

	t.Run("permanent_error", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			return &PermanentError{Err: errors.New("permanent error")}
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		config.SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		retry := NewRetry(config)
		err := retry.Execute(ctx, work)
		require.Error(t, err)
		require.Equal(t, 1, attempts)

		unwrappedErr := errors.Unwrap(err)
		require.NotNil(t, unwrappedErr)
		require.Equal(t, "permanent error", unwrappedErr.Error())
	})

	t.Run("default_backoff", func(t *testing.T) {
		config := &configv1.RetryConfig{}
		retry := NewRetry(config)
		require.Equal(t, time.Second, retry.config.GetBaseBackoff().AsDuration())
		require.Equal(t, 30*time.Second, retry.config.GetMaxBackoff().AsDuration())
	})

	t.Run("backoff_capping", func(t *testing.T) {
		config := &configv1.RetryConfig{}
		config.SetBaseBackoff(durationpb.New(2 * time.Second))
		config.SetMaxBackoff(durationpb.New(5 * time.Second))
		retry := NewRetry(config)
		// backoff(1) should be around 4s (2s * 2) with jitter
		// 4s * 0.8 = 3.2s, 4s * 1.2 = 4.8s. InDelta 1s covers this.
		require.InDelta(t, float64(4*time.Second), float64(retry.backoff(1)), float64(1*time.Second))

		// backoff(2) hits the cap (2s * 4 = 8s > 5s).
		// Current implementation returns maxBackoff exactly when capped?
		// Let's assume yes based on code reading.
		require.Equal(t, 5*time.Second, retry.backoff(2))
	})

	t.Run("no_wait_after_last_attempt", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			return errors.New("persistent error")
		}

		retries := int32(2)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		// Set a long backoff to make it obvious if we wait
		config.SetBaseBackoff(durationpb.New(100 * time.Millisecond))
		retry := NewRetry(config)

		start := time.Now()
		err := retry.Execute(ctx, work)
		elapsed := time.Since(start)

		require.Error(t, err)
		require.Equal(t, 3, attempts)

		require.Less(t, elapsed, 600*time.Millisecond, "should not wait after the last attempt")
	})

	t.Run("context_cancellation_during_backoff", func(t *testing.T) {
		// This test verifies that if the context is cancelled while waiting for backoff,
		// the retry logic returns immediately with the context error.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var attempts int
		work := func(_ context.Context) error {
			attempts++
			return errors.New("error")
		}

		retries := int32(5)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		// Set a very long backoff so the test would timeout if we didn't respect cancellation
		config.SetBaseBackoff(durationpb.New(1 * time.Hour))
		retry := NewRetry(config)

		// Start a goroutine to cancel the context after a short delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := retry.Execute(ctx, work)
		elapsed := time.Since(start)

		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
		// Should return shortly after 100ms
		require.Less(t, elapsed, 2*time.Second, "should return immediately on context cancellation")
		// Should have attempted at least once
		require.GreaterOrEqual(t, attempts, 1)
	})
}
