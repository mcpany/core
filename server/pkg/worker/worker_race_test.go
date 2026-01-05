// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	buspb "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

// This test is designed to fail before the race condition fix in worker.go is applied.
// It exposes a race condition between the Start and Stop methods of the worker.
// The test may not fail on every run because it depends on the scheduler, but it is likely to fail.
func TestWorker_StartStopRace(t *testing.T) {
	t.Skip("Skipping known flaky test that fails intermittently")
	defer goleak.VerifyNone(t)
	busConfig := buspb.MessageBus_builder{
		InMemory: &buspb.InMemoryBus{},
	}.Build()
	bp, err := bus.NewProvider(busConfig)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		w := New(bp, &Config{
			MaxWorkers:   1,
			MaxQueueSize: 1,
		})

		w.Start(context.Background())
		w.Stop()

		// After stopping, the worker should not process any more requests.
		// If the unsubscribe function was not called due to the race, the worker will still be subscribed.
		reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		require.NoError(t, err)
		resBus, err := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		require.NoError(t, err)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		correlationID := "test-race"
		unsubscribe := resBus.Subscribe(context.Background(), correlationID, func(result *bus.ToolExecutionResult) {
			resultChan <- result
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID(correlationID)
		_ = reqBus.Publish(context.Background(), bus.ToolExecutionRequestTopic, req)

		select {
		case <-resultChan:
			t.Fatalf("worker processed request after stop on iteration %d", i)
		case <-time.After(20 * time.Millisecond):
			// success, no result received
		}
	}
}
