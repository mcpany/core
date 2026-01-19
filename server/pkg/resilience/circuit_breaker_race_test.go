// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCircuitBreaker_Flapping_Race(t *testing.T) {
	ctx := context.Background()
	consecutiveFailures := int32(1)
	// We want to allow 2 concurrent probes
	halfOpenRequests := int32(2)

	config := &configv1.CircuitBreakerConfig{}
	config.SetConsecutiveFailures(consecutiveFailures)
	config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
	config.SetHalfOpenRequests(halfOpenRequests)
	cb := NewCircuitBreaker(config)

	// 1. Open the circuit
	_ = cb.Execute(ctx, func(_ context.Context) error { return errors.New("error") })
	require.Equal(t, StateOpen, cb.state)

	// 2. Wait for OpenDuration to pass
	time.Sleep(20 * time.Millisecond)

	// 3. Launch 2 concurrent requests
	// Both should see "HalfOpen" state (or transition to it).
	// We use channels to synchronize them to ensure they are both "in flight" inside Execute.

	// Since we can't pause execution inside `Execute`, we rely on `work` function blocking.

	startWork := make(chan struct{})
	finishWorkSuccess := make(chan struct{})
	finishWorkFailure := make(chan struct{})
	// Buffer channel to avoid blocking sends
	started := make(chan struct{}, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	// Request 1: Succeeds
	go func() {
		defer wg.Done()
		_ = cb.Execute(ctx, func(_ context.Context) error {
			started <- struct{}{}
			<-startWork // Wait until both are running
			<-finishWorkSuccess // Wait for signal to finish
			return nil
		})
	}()

	// Request 2: Fails
	go func() {
		defer wg.Done()
		_ = cb.Execute(ctx, func(_ context.Context) error {
			started <- struct{}{}
			<-startWork
			<-finishWorkFailure
			return errors.New("error")
		})
	}()

	// Wait for both goroutines to enter `Execute` and pass the HalfOpen check
	// The fact they reached `started` means they are inside `work`, so they passed validation.
	<-started
	<-started

	// Start both
	close(startWork)

	// Now both are running `work`.
	// State should be HalfOpen.
	// Verify state? Race condition to check state here.

	// Let Request 1 finish (Succeed).
	close(finishWorkSuccess)

	// Wait a tiny bit to ensure Request 1 processes success and Closes the breaker.
	// This sleep is acceptable as we are waiting for an async state transition.
	time.Sleep(10 * time.Millisecond)

	// At this point, Breaker should be Closed.
	// And Failures = 0.

	if cb.getState() != StateClosed {
		t.Logf("State is not Closed after success! State: %v", cb.getState())
	}

	// Let Request 2 finish (Fail).
	close(finishWorkFailure)

	wg.Wait()

	// Check final state.
	// If the bug exists, Request 2's failure will be counted against the now-Closed breaker.
	// Since ConsecutiveFailures = 1, it will trip to Open.
	// Ideally, it should remain Closed because the service recovered (proven by Req 1).

	require.Equal(t, StateClosed, cb.state, "Circuit breaker should remain closed after recovery, despite concurrent probe failure")
}
