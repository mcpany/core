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

// UpstreamWorker is a background worker that handles tool execution requests. It.
//
// Summary: is a background worker that handles tool execution requests. It.
type UpstreamWorker struct {
	bus         *bus.Provider
	toolManager tool.ManagerInterface
	wg          sync.WaitGroup
}

// NewUpstreamWorker creates a new UpstreamWorker.
//
// Summary: creates a new UpstreamWorker.
//
// Parameters:
//   - bus: *bus.Provider. The bus.
//   - toolManager: tool.ManagerInterface. The toolManager.
//
// Returns:
//   - *UpstreamWorker: The *UpstreamWorker.
func NewUpstreamWorker(bus *bus.Provider, toolManager tool.ManagerInterface) *UpstreamWorker {
	return &UpstreamWorker{
		bus:         bus,
		toolManager: toolManager,
	}
}

// Start launches the worker in a new goroutine. It subscribes to tool execution.
//
// Summary: launches the worker in a new goroutine. It subscribes to tool execution.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   None.
func (w *UpstreamWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	log := logging.GetLogger().With("component", "UpstreamWorker")
	log.Info("Upstream worker started")

	requestBus, _ := bus.GetBus[*bus.ToolExecutionRequest](w.bus, bus.ToolExecutionRequestTopic)
	resultBus, _ := bus.GetBus[*bus.ToolExecutionResult](w.bus, bus.ToolExecutionResultTopic)

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
//
// Summary: waits for the worker to stop.
//
// Parameters:
//   None.
//
// Returns:
//   None.
func (w *UpstreamWorker) Stop() {
	w.wg.Wait()
}
