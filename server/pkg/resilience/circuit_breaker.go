// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// State represents the current state of the circuit breaker.
//
// Summary: represents the current state of the circuit breaker.
type State int32

const (
	// StateClosed represents the state where the circuit breaker allows requests to pass through.
	StateClosed State = iota
	// StateOpen represents the state where the circuit breaker blocks requests immediately.
	StateOpen
	// StateHalfOpen represents the state where the circuit breaker allows a limited number of requests to test if the service has recovered.
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern. It prevents the.
//
// Summary: implements the circuit breaker pattern. It prevents the.
type CircuitBreaker struct {
	mutex sync.Mutex

	state        State // Accessed using atomics for read optimization
	failures     int32 // Accessed using atomics
	openTime     time.Time
	halfOpenHits int

	config *configv1.CircuitBreakerConfig
}

// NewCircuitBreaker creates a new CircuitBreaker with the given configuration.
//
// Summary: creates a new CircuitBreaker with the given configuration.
//
// Parameters:
//   - config: *configv1.CircuitBreakerConfig. The config.
//
// Returns:
//   - *CircuitBreaker: The *CircuitBreaker.
func NewCircuitBreaker(config *configv1.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute runs the provided work function. If the circuit breaker is open, it.
//
// Summary: runs the provided work function. If the circuit breaker is open, it.
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
func (cb *CircuitBreaker) Execute(ctx context.Context, work func(context.Context) error) error {
	originState := StateClosed

	// Optimization: Optimistically check if Closed without lock.
	// This covers the "Happy Path" (most common case).
	if cb.getState() != StateClosed {
		// Slow path: acquire lock to check Open/HalfOpen state
		cb.mutex.Lock()

		// Re-check state under lock
		currentState := cb.getState()
		originState = currentState

		if currentState == StateOpen {
			if time.Since(cb.openTime) > cb.config.GetOpenDuration().AsDuration() {
				cb.setState(StateHalfOpen)
				cb.halfOpenHits = 0
				currentState = StateHalfOpen
				originState = StateHalfOpen
			} else {
				cb.mutex.Unlock()
				return &CircuitBreakerOpenError{}
			}
		}

		if currentState == StateHalfOpen {
			if cb.halfOpenHits >= int(cb.config.GetHalfOpenRequests()) {
				cb.mutex.Unlock()
				return &CircuitBreakerOpenError{}
			}
			cb.halfOpenHits++
		}

		cb.mutex.Unlock()
	}

	err := work(ctx)
	if err != nil {
		var permanentErr *PermanentError
		if errors.As(err, &permanentErr) {
			return err
		}

		// ⚡ BOLT: Removed mutex lock around onFailure to reduce contention.
		// Randomized Selection from Top 5 High-Impact Targets
		cb.onFailure(originState)
		return err
	}

	// Optimization: If the circuit is closed and there are no recorded failures,
	// we can skip acquiring the lock. This covers the "Happy Path" where
	// everything is working normally.
	if cb.getState() == StateClosed && atomic.LoadInt32(&cb.failures) == 0 {
		return nil
	}

	// ⚡ BOLT: Removed mutex lock around onSuccess to reduce contention.
	// Randomized Selection from Top 5 High-Impact Targets
	cb.onSuccess(originState)
	return nil
}

// getState reads the state atomically.
func (cb *CircuitBreaker) getState() State {
	return State(atomic.LoadInt32((*int32)(&cb.state)))
}

// setState updates the state atomically. Caller must hold the mutex.
func (cb *CircuitBreaker) setState(newState State) {
	atomic.StoreInt32((*int32)(&cb.state), int32(newState))
}

func (cb *CircuitBreaker) onSuccess(originState State) {
	// Optimization: Fast path for Closed state.
	// If state is Closed, we just need to reset failures.
	if cb.getState() == StateClosed {
		atomic.StoreInt32(&cb.failures, 0)
		return
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.getState() == StateHalfOpen {
		if originState != StateHalfOpen {
			return
		}
		cb.setState(StateClosed)
		cb.halfOpenHits = 0
	}
	atomic.StoreInt32(&cb.failures, 0)
}

func (cb *CircuitBreaker) onFailure(originState State) {
	currentState := cb.getState()

	// Optimization: Fast path for Closed state (most common).
	// We increment failures atomically. Only if we hit threshold do we lock.
	// We must also check that we STARTED in Closed state. If we started in HalfOpen,
	// we need to run the slow path to check for flapping (HalfOpen -> Closed transition).
	if currentState == StateClosed && originState == StateClosed {
		newFailures := atomic.AddInt32(&cb.failures, 1)
		if newFailures >= cb.config.GetConsecutiveFailures() {
			cb.mutex.Lock()
			defer cb.mutex.Unlock()

			// Re-check state to handle races
			if cb.getState() == StateClosed {
				cb.setState(StateOpen)
				cb.openTime = time.Now()
			}
		}
		return
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Re-read state under lock
	currentState = cb.getState()

	if currentState == StateOpen {
		return
	}

	// If the request started in HalfOpen state but the breaker is now Closed,
	// it means another concurrent probe succeeded. In this case, we ignore
	// this failure to prevent flapping (Closing then immediately Opening).
	if originState == StateHalfOpen && currentState == StateClosed {
		return
	}

	if currentState == StateHalfOpen {
		// Only trip the breaker if the failure corresponds to a probe (started in HalfOpen state).
		// If the request started in Closed state (e.g., a slow request from before the break),
		// we ignore it to allow the current probes to complete.
		if originState != StateHalfOpen {
			return
		}
		cb.setState(StateOpen)
		cb.openTime = time.Now()
		return
	}

	newFailures := atomic.AddInt32(&cb.failures, 1)
	if newFailures >= cb.config.GetConsecutiveFailures() {
		cb.setState(StateOpen)
		cb.openTime = time.Now()
	}
}

// CircuitBreakerOpenError is returned when the circuit breaker is in the Open state.
//
// Summary: is returned when the circuit breaker is in the Open state.
type CircuitBreakerOpenError struct{}

// Error returns the error message for a CircuitBreakerOpenError.
//
// Summary: returns the error message for a CircuitBreakerOpenError.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (e *CircuitBreakerOpenError) Error() string {
	return "circuit breaker is open"
}
