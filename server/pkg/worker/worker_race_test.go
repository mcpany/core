// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"testing"
	"time"

	buspb "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

// This test ensures that the race condition between the Start and Stop methods of the worker is fixed.
// It verifies that Stop() waits for the worker to fully start (and subscribe) before proceeding to unsubscribe.
func TestWorker_StartStopRace(t *testing.T) {
	defer func() {
		// Give some time for subscriber goroutines to exit
		time.Sleep(100 * time.Millisecond)
		goleak.VerifyNone(t)
	}()
	busConfig := buspb.MessageBus_builder{
		InMemory: &buspb.InMemoryBus{},
	}.Build()
	bp, err := bus.NewProvider(busConfig)
	require.NoError(t, err)

	for i := 0; i < 50; i++ {
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
