package resilience

import (
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestManager(t *testing.T) {
	t.Run("execute_with_retry", func(t *testing.T) {
		var attempts int
		work := func() error {
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
		err := manager.Execute(work)
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
		_ = manager.Execute(func() error { return errors.New("error") })
		_ = manager.Execute(func() error { return errors.New("error") })

		// Third request should be blocked
		err := manager.Execute(func() error { return nil })
		require.Error(t, err)
		require.IsType(t, &CircuitBreakerOpenError{}, err)
	})

	t.Run("execute_with_retry_and_circuit_breaker", func(t *testing.T) {
		var attempts int
		work := func() error {
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
		err := manager.Execute(work)
		require.Error(t, err)
		require.Equal(t, 2, attempts)

		// The circuit breaker should now be open
		err = manager.Execute(work)
		require.Error(t, err)
		require.IsType(t, &CircuitBreakerOpenError{}, err)
		require.Equal(t, 2, attempts)
	})

	t.Run("nil_config", func(t *testing.T) {
		manager := NewManager(nil)
		err := manager.Execute(func() error { return nil })
		require.NoError(t, err)
	})
}
