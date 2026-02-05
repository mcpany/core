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

func TestCircuitBreaker_Backlog_Scenario(t *testing.T) {
	ctx := context.Background()
	consecutiveFailures := int32(1)
	openDuration := 50 * time.Millisecond
	config := &configv1.CircuitBreakerConfig{}
	config.SetConsecutiveFailures(consecutiveFailures)
	config.SetOpenDuration(durationpb.New(openDuration))
	config.SetHalfOpenRequests(1) // Allow 1 probe

	cb := NewCircuitBreaker(config)

	// 1. Start a slow request (Request A) while Closed.
	// It will block on a channel.
	startA := make(chan struct{})
	finishA := make(chan error)
	go func() {
		_ = cb.Execute(ctx, func(_ context.Context) error {
			close(startA)
			return <-finishA
		})
	}()

	<-startA // Request A is now "in flight" (past the lock)

	// 2. Trigger a failure (Request B) to open the breaker.
	err := cb.Execute(ctx, func(_ context.Context) error { return errors.New("fail B") })
	require.Error(t, err)
	require.Equal(t, StateOpen, cb.getState())

	// 3. Wait for OpenDuration to expire.
	time.Sleep(openDuration + 10*time.Millisecond)

	// 4. Start a probe request (Request C).
	// This should transition state to HalfOpen.
	// We want this probe to eventually succeed, but let's pause it too to control timing.
	startC := make(chan struct{})
	finishC := make(chan error)

	// We run this in goroutine because Execute might block or we want to control execution.
	// Actually, the first request after wait should enter HalfOpen.
	go func() {
		_ = cb.Execute(ctx, func(_ context.Context) error {
			close(startC)
			return <-finishC
		})
	}()

	<-startC
	require.Equal(t, StateHalfOpen, cb.getState())

	// 5. Now, Request A (the old one) fails.
	finishA <- errors.New("fail A")

	// Give it a moment to process onFailure.
	time.Sleep(10 * time.Millisecond)

	// 6. Check state.
	// The old request failure (originState=Closed) should be ignored in HalfOpen state.
	// So state should remain StateHalfOpen.
	state := cb.getState()
	require.Equal(t, StateHalfOpen, state, "Old request failure should not trip the breaker from HalfOpen")

	// Cleanup C
	finishC <- nil
}
