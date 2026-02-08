package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
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
		require.Equal(t, int32(0), cb.failures)
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

// TestCircuitBreaker_ZombieSuccess_ClosesBreaker checks if a success from a stale request
// (started when Closed) incorrectly closes the breaker when it is HalfOpen.
func TestCircuitBreaker_ZombieSuccess_ClosesBreaker(t *testing.T) {
	consecutiveFailures := int32(2)
	halfOpenRequests := int32(1)
	openDuration := 50 * time.Millisecond

	config := &configv1.CircuitBreakerConfig{}
	config.SetConsecutiveFailures(consecutiveFailures)
	config.SetOpenDuration(durationpb.New(openDuration))
	config.SetHalfOpenRequests(halfOpenRequests)
	cb := NewCircuitBreaker(config)

	var wg sync.WaitGroup
	ctx := context.Background()

	// 1. Start a slow request (A) while Closed.
	// It will sleep long enough to span across Open and HalfOpen states.
	zombieStarted := make(chan struct{})
	zombieSuccess := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = cb.Execute(ctx, func(_ context.Context) error {
			close(zombieStarted)
			<-zombieSuccess // Wait until we signal it to finish
			return nil      // Success!
		})
	}()

	// Wait for A to start and register "Closed" as origin state.
	<-zombieStarted

	// 2. Fail requests to trip the breaker.
	for i := 0; i < int(consecutiveFailures); i++ {
		_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("fail") })
	}
	require.Equal(t, StateOpen, cb.getState(), "Breaker should be Open")

	// 3. Wait for Open duration to expire.
	time.Sleep(openDuration + 10*time.Millisecond)

	// 4. Trigger HalfOpen state with a new request (Probe).
	// We trigger transition to HalfOpen using a probe that we block.
	probeStarted := make(chan struct{})
	probeProceed := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = cb.Execute(ctx, func(_ context.Context) error {
			// We are now in HalfOpen state inside the lock (partially) or just executed logic.
			// The state transition happens BEFORE work is called.
			close(probeStarted)
			<-probeProceed
			return nil
		})
	}()

	// Wait for Probe to start and ensure transition state to HalfOpen.
	<-probeStarted
	require.Equal(t, StateHalfOpen, cb.getState(), "Breaker should be HalfOpen")

	// 5. Now let the Zombie request (A) succeed.
	close(zombieSuccess)

	// We need to wait for A to finish processing onSuccess.
	// Since onSuccess happens after work returns, and we just unblocked work,
	// we need a small sleep or another synchronization mechanism.
	time.Sleep(10 * time.Millisecond)

	// 6. Check state.
	// If the bug exists, the Zombie success will Close the breaker.
	// BUT we still have the Probe running!
	// We want the breaker to remain HalfOpen until the *Probe* succeeds.
	state := cb.getState()

	// Clean up
	close(probeProceed)
	wg.Wait()

	assert.Equal(t, StateHalfOpen, state, "Breaker should remain HalfOpen after zombie success")
}
