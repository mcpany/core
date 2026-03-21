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

func TestManager(t *testing.T) {
	ctx := context.Background()
	t.Run("execute_with_retry", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("transient error")
			}
			return nil
		}

		retries := int32(3)
		config := &configv1.ResilienceConfig{}
		config.SetRetryPolicy(&configv1.RetryConfig{})
		config.GetRetryPolicy().SetNumberOfRetries(retries)
		config.GetRetryPolicy().SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		manager := NewManager(config)
		err := manager.Execute(ctx, work)
		require.NoError(t, err)
		require.Equal(t, 3, attempts)
	})

	t.Run("execute_with_circuit_breaker", func(t *testing.T) {
		consecutiveFailures := int32(2)
		config := &configv1.ResilienceConfig{}
		config.SetCircuitBreaker(&configv1.CircuitBreakerConfig{})
		config.GetCircuitBreaker().SetConsecutiveFailures(consecutiveFailures)
		config.GetCircuitBreaker().SetOpenDuration(durationpb.New(10 * time.Second))
		manager := NewManager(config)

		// Fail twice to open the circuit
		_ = manager.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
		_ = manager.Execute(ctx, func(_ context.Context) error { return errors.New("error") })

		// Third request should be blocked
		err := manager.Execute(ctx, func(_ context.Context) error { return nil })
		require.Error(t, err)
		require.IsType(t, &CircuitBreakerOpenError{}, err)
	})

	t.Run("execute_with_retry_and_circuit_breaker", func(t *testing.T) {
		var attempts int
		work := func(_ context.Context) error {
			attempts++
			return errors.New("persistent error")
		}

		retries := int32(3)
		consecutiveFailures := int32(2)
		config := &configv1.ResilienceConfig{}
		config.SetRetryPolicy(&configv1.RetryConfig{})
		config.GetRetryPolicy().SetNumberOfRetries(retries)
		config.GetRetryPolicy().SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		config.SetCircuitBreaker(&configv1.CircuitBreakerConfig{})
		config.GetCircuitBreaker().SetConsecutiveFailures(consecutiveFailures)
		config.GetCircuitBreaker().SetOpenDuration(durationpb.New(10 * time.Second))
		manager := NewManager(config)

		// The circuit breaker will open after 2 failures, halting further retries.
		err := manager.Execute(ctx, work)
		require.Error(t, err)
		require.Equal(t, 2, attempts)

		// The circuit breaker should now be open
		err = manager.Execute(ctx, work)
		require.Error(t, err)
		require.IsType(t, &CircuitBreakerOpenError{}, err)
		require.Equal(t, 2, attempts)
	})

	t.Run("nil_config", func(t *testing.T) {
		manager := NewManager(nil)
		err := manager.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)
	})

    t.Run("empty_config_returns_nil", func(t *testing.T) {
		config := &configv1.ResilienceConfig{}
		manager := NewManager(config)
		require.Nil(t, manager)
	})

     t.Run("manager_nil_check", func(t *testing.T) {
        var manager *Manager
		err := manager.Execute(ctx, func(_ context.Context) error { return nil })
		require.NoError(t, err)
	})
}

func TestManager_Execute_WithTimeout(t *testing.T) {
    ctx := context.Background()
	t.Run("Timeout_triggers", func(t *testing.T) {
		config := &configv1.ResilienceConfig{}
		config.SetTimeout(durationpb.New(50 * time.Millisecond))
		manager := NewManager(config)

		err := manager.Execute(ctx, func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		})
		require.Error(t, err)
        require.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("Timeout_does_not_trigger", func(t *testing.T) {
		config := &configv1.ResilienceConfig{}
		config.SetTimeout(durationpb.New(100 * time.Millisecond))
		manager := NewManager(config)

		err := manager.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		require.NoError(t, err)
	})
}
