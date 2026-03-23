// Copyright 2026 Author(s) of MCP Any
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

func TestNewRetry_NilConfig_Works(t *testing.T) {
	// Should not panic
	retry := NewRetry(nil)
	require.NotNil(t, retry)
	require.NotNil(t, retry.config)

	// Should work with defaults (1s base, 30s max, 0 retries by default)
	require.Equal(t, time.Second, retry.config.GetBaseBackoff().AsDuration())
	require.Equal(t, 30*time.Second, retry.config.GetMaxBackoff().AsDuration())
	require.Equal(t, int32(0), retry.config.GetNumberOfRetries())

	// Execute should work
	ctx := context.Background()
	work := func(_ context.Context) error {
		return errors.New("fail")
	}
	err := retry.Execute(ctx, work)
	require.Error(t, err)
}

func TestRetry_Backoff_Overflow(t *testing.T) {
	config := &configv1.RetryConfig{}
	config.SetBaseBackoff(durationpb.New(time.Second))
	config.SetMaxBackoff(durationpb.New(100 * time.Second))
	retry := NewRetry(config)

	// Case 1: Attempt 30.
	dur := retry.backoff(30)
	require.Equal(t, 100*time.Second, dur, "Backoff(30) should be capped at MaxBackoff")

	// Case 2: Attempt 63.
	dur = retry.backoff(63)
	require.Equal(t, 100*time.Second, dur, "Backoff(63) should be capped at MaxBackoff")

	// Case 3: Attempt 100 (way over 64)
	dur = retry.backoff(100)
	require.Equal(t, 100*time.Second, dur, "Backoff(100) should be capped at MaxBackoff")

    // Case 4: Negative attempt
    dur = retry.backoff(-1)
    require.Equal(t, time.Second, dur, "Backoff(-1) should return BaseBackoff")
}

func TestRetry_Backoff_BaseZero(t *testing.T) {
    config := &configv1.RetryConfig{}
    config.SetBaseBackoff(durationpb.New(0))
    config.SetMaxBackoff(durationpb.New(10 * time.Second))
    retry := NewRetry(config)

    dur := retry.backoff(1)
    require.Equal(t, time.Duration(0), dur)
}

func TestRetry_ContextCancellation(t *testing.T) {
    // Test context cancelled before execution
    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    retry := NewRetry(nil) // default config
    err := retry.Execute(ctx, func(context.Context) error { return nil })
    require.ErrorIs(t, err, context.Canceled)

    // Test context cancelled during backoff
    ctx2, cancel2 := context.WithCancel(context.Background())

    config := &configv1.RetryConfig{}
    config.SetNumberOfRetries(1)
    config.SetBaseBackoff(durationpb.New(500 * time.Millisecond)) // Long enough to catch
    retry2 := NewRetry(config)

    start := time.Now()
    go func() {
        time.Sleep(100 * time.Millisecond)
        cancel2()
    }()

    err2 := retry2.Execute(ctx2, func(ctx context.Context) error {
        return errors.New("fail")
    })

    require.ErrorIs(t, err2, context.Canceled)
    elapsed := time.Since(start)
    require.Less(t, elapsed, 400*time.Millisecond, "Should have returned early due to cancellation")
}

func TestCircuitBreaker_HalfOpenLimit(t *testing.T) {
	config := &configv1.CircuitBreakerConfig{}
	config.SetHalfOpenRequests(1)
    config.SetOpenDuration(durationpb.New(time.Millisecond))
	cb := NewCircuitBreaker(config)

    // Force state to HalfOpen
    cb.setState(StateHalfOpen)
    cb.halfOpenHits = 0

    ctx := context.Background()

    // First request should pass
    err := cb.Execute(ctx, func(_ context.Context) error { return nil })
    require.NoError(t, err)

    // Check if it transitioned to closed on success
    require.Equal(t, StateClosed, cb.getState())

    // Reset to HalfOpen to test limit
    cb.setState(StateHalfOpen)
    cb.halfOpenHits = 0

    // To test limit, we need concurrent requests or one request failing/succeeding but not changing state?
    // Wait, Execute calls onSuccess/onFailure which changes state.
    // If we want to hit limit, we need multiple checks BEFORE one finishes?
    // Or we rely on `cb.Execute` logic:
    /*
		if currentState == StateHalfOpen {
			if cb.halfOpenHits >= int(cb.config.GetHalfOpenRequests()) {
				cb.mutex.Unlock()
				return &CircuitBreakerOpenError{}
			}
			cb.halfOpenHits++
		}
    */
    // We need to trigger this without transitioning out of HalfOpen immediately?
    // But `Execute` calls `work` then `onSuccess` or `onFailure`.
    // If we block inside `work`, `halfOpenHits` is incremented.
    // Then another `Execute` call comes in.

    ready := make(chan struct{})
    block := make(chan struct{})

    go func() {
        cb.Execute(ctx, func(_ context.Context) error {
            close(ready)
            <-block
            return nil
        })
    }()

    <-ready
    // Now first request is in progress. halfOpenHits should be 1.
    // Limit is 1.

    // Second request should be rejected
    err = cb.Execute(ctx, func(_ context.Context) error { return nil })
    require.Error(t, err)
    require.IsType(t, &CircuitBreakerOpenError{}, err)

    close(block)
}

func TestCircuitBreaker_OnFailure_WhenOpen(t *testing.T) {
    config := &configv1.CircuitBreakerConfig{}
    cb := NewCircuitBreaker(config)

    cb.setState(StateOpen)

    // Should return early
    cb.onFailure(StateClosed)

    // State should remain Open
    require.Equal(t, StateOpen, cb.getState())
}
