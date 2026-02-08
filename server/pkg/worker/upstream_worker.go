// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/tool"
)

// UpstreamWorker is a background worker that handles tool execution requests. It
// listens for ToolExecutionRequest messages on the event bus, uses the
// tool manager to execute the requested tool, and then publishes the outcome as
// a ToolExecutionResult message.
type UpstreamWorker struct {
	bus         *bus.Provider
	toolManager tool.ManagerInterface
	wg          sync.WaitGroup
}

// NewUpstreamWorker creates a new UpstreamWorker.
//
// Parameters:
//   - bus: The event bus used for receiving requests and publishing results.
//   - toolManager: The tool manager that will handle the actual tool execution.
//
// Returns:
//   - *UpstreamWorker: A new upstream worker.
func NewUpstreamWorker(bus *bus.Provider, toolManager tool.ManagerInterface) *UpstreamWorker {
	return &UpstreamWorker{
		bus:         bus,
		toolManager: toolManager,
	}
}

// Start launches the worker in a new goroutine. It subscribes to tool execution
// requests on the event bus and will continue to process them until the
// provided context is canceled.
//
// Parameters:
//   - ctx: The context that controls the lifecycle of the worker.
func (w *UpstreamWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	log := logging.GetLogger().With("component", "UpstreamWorker")
	log.Info("Upstream worker started")

	requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](w.bus, bus.ToolExecutionRequestTopic)
	if err != nil {
		log.Error("Failed to get request bus", "error", err)
		w.wg.Done()
		return
	}

	resultBus, err := bus.GetBus[*bus.ToolExecutionResult](w.bus, bus.ToolExecutionResultTopic)
	if err != nil {
		log.Error("Failed to get result bus", "error", err)
		w.wg.Done()
		return
	}

	unsubscribe := requestBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
		start := time.Now()
		metrics.IncrCounter([]string{"worker", "upstream", "request", "total"}, 1)
		defer metrics.MeasureSince([]string{"worker", "upstream", "request", "latency"}, start)
		log.Info("Received tool execution request", "tool", req.ToolName, "correlationID", req.CorrelationID())
		result, err := w.toolManager.ExecuteTool(req.Context, &tool.ExecutionRequest{
			ToolName:   req.ToolName,
			ToolInputs: req.ToolInputs,
		})

		var resultBytes json.RawMessage
		if result != nil {
			var marshalErr error
			resultBytes, marshalErr = json.Marshal(result)
			if marshalErr != nil {
				log.Error("Failed to marshal tool execution result", "error", marshalErr)
				err = marshalErr
			}
		}

		res := &bus.ToolExecutionResult{
			Result: resultBytes,
			Error:  err,
		}
		if err != nil {
			metrics.IncrCounter([]string{"worker", "upstream", "request", "error"}, 1)
		} else {
			metrics.IncrCounter([]string{"worker", "upstream", "request", "success"}, 1)
		}
		res.SetCorrelationID(req.CorrelationID())
		if err := resultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
			log.Error("Failed to publish tool execution result", "error", err)
		}
	})

	go func() {
		defer w.wg.Done()
		<-ctx.Done()
		log.Info("Upstream worker stopping")
		unsubscribe()
	}()
}

// Stop waits for the worker to stop.
func (w *UpstreamWorker) Stop() {
	w.wg.Wait()
}
