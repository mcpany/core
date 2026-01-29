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

func TestCircuitBreaker_HalfOpenLimitBug(t *testing.T) {
	ctx := context.Background()
	consecutiveFailures := int32(1)
	halfOpenRequests := int32(1)
	config := &configv1.CircuitBreakerConfig{}
	config.SetConsecutiveFailures(consecutiveFailures)
	config.SetOpenDuration(durationpb.New(10 * time.Millisecond))
	config.SetHalfOpenRequests(halfOpenRequests)
	cb := NewCircuitBreaker(config)

	// 1. Open the circuit
	err := cb.Execute(ctx, func(_ context.Context) error { return errors.New("fail") })
	require.Error(t, err)
	require.Equal(t, StateOpen, cb.state)

	// 2. Wait for OpenDuration to pass
	time.Sleep(20 * time.Millisecond)

	// 3. Req 1: Transition to HalfOpen and hold
	req1Started := make(chan struct{})
	req1Block := make(chan struct{})
	req1Done := make(chan error)

	go func() {
		err := cb.Execute(ctx, func(_ context.Context) error {
			close(req1Started)
			<-req1Block
			return nil
		})
		req1Done <- err
	}()

	// Wait for Req 1 to be inside work (state should be HalfOpen)
	select {
	case <-req1Started:
	case <-time.After(1 * time.Second):
		t.Fatal("Req 1 did not start")
	}

	// Verify state is HalfOpen
	// Note: We need to access state. Tests are in same package so we can access private fields.
	require.Equal(t, StateHalfOpen, cb.getState())

	// 4. Req 2: Should be rejected if limit is 1
	// With the bug, halfOpenHits is 0 (because Req 1 didn't increment it during transition),
	// so Req 2 increments it to 1 and runs.
	req2Ran := false
	err = cb.Execute(ctx, func(_ context.Context) error {
		req2Ran = true
		return nil
	})

	// Cleanup Req 1
	close(req1Block)
	<-req1Done

	// Assertions
	require.False(t, req2Ran, "Second request should not have run")
	require.Error(t, err)
	require.IsType(t, &CircuitBreakerOpenError{}, err)
}
