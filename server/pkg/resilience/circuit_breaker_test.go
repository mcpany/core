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

func TestCircuitBreaker(t *testing.T) {
	ctx := context.Background()
	t.Run("closed_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		cb := NewCircuitBreaker(config)
		err := cb.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)
		require.Equal(t, StateClosed, cb.state)
	})

	t.Run("open_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Second))
		cb := NewCircuitBreaker(config)

		// Fail twice to open the circuit
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })

		require.Equal(t, StateOpen, cb.state)

		// Third request should be blocked
		err := cb.Execute(ctx, func(_ context.Context) error { return nil })
		require.Error(t, err)
		require.IsType(t, &CircuitBreakerOpenError{}, err)
	})

	t.Run("half_open_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
		config.SetHalfOpenRequests(1)
		cb := NewCircuitBreaker(config)

		// Fail twice to open the circuit
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })

		// Wait for the open duration to elapse
		time.Sleep(15 * time.Millisecond)

		// First request in half-open state should be allowed
		err := cb.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)
		require.Equal(t, StateClosed, cb.state)
	})

	t.Run("half_open_to_open_state", func(t *testing.T) {
		consecutiveFailures := int32(2)
		halfOpenRequests := int32(1)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
		config.SetHalfOpenRequests(halfOpenRequests)
		cb := NewCircuitBreaker(config)

		// Fail twice to open the circuit
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })

		// Wait for the open duration to elapse
		time.Sleep(15 * time.Millisecond)

		// First request in half-open state should be allowed, but fail
		err := cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
		require.Error(t, err)
		require.Equal(t, StateOpen, cb.state)
	})

	t.Run("permanent_error", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		cb := NewCircuitBreaker(config)

		// A permanent error should not affect the circuit breaker state
		err := cb.Execute(ctx, func(_ context.Context) error { return &PermanentError{Err: errors.New("permanent error")} })
		require.Error(t, err)
		require.Equal(t, StateClosed, cb.state)
		require.Equal(t, 0, cb.failures)
	})

	t.Run("half_open_to_closed_on_success", func(t *testing.T) {
		consecutiveFailures := int32(1)
		halfOpenRequests := int32(2)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
		config.SetHalfOpenRequests(halfOpenRequests)
		cb := NewCircuitBreaker(config)

		// Fail once to open the circuit
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })

		// Wait for the open duration to elapse
		time.Sleep(15 * time.Millisecond)

		// The circuit is now half-open.

		// First and second requests in half-open state should be allowed
		err := cb.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)
		err = cb.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)

		// The circuit should now be closed.
		require.Equal(t, StateClosed, cb.state)

		// Third request should also be allowed.
		err = cb.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)
	})

	t.Run("half_open_to_open_on_failure", func(t *testing.T) {
		consecutiveFailures := int32(1)
		halfOpenRequests := int32(1)
		config := &configv1.CircuitBreakerConfig{}
		config.SetConsecutiveFailures(consecutiveFailures)
		config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
		config.SetHalfOpenRequests(halfOpenRequests)
		cb := NewCircuitBreaker(config)

		// Fail once to open the circuit
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })

		// Wait for the open duration to elapse
		time.Sleep(15 * time.Millisecond)

		// The circuit is now half-open.

		// First request in half-open state should be allowed, but fail
		err := cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
		require.Error(t, err)

		// The circuit should now be open.
		require.Equal(t, StateOpen, cb.state)
	})
}
