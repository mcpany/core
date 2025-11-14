/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
)

// UpstreamWorker is a background worker that handles tool execution requests. It
// listens for ToolExecutionRequest messages on the event bus, uses the
// tool manager to execute the requested tool, and then publishes the outcome as
// a ToolExecutionResult message.
type UpstreamWorker struct {
	bus         *bus.BusProvider
	toolManager tool.ToolManagerInterface
}

// NewUpstreamWorker creates a new UpstreamWorker.
//
// bus is the event bus used for receiving requests and publishing results.
// toolManager is the tool manager that will handle the actual tool execution.
func NewUpstreamWorker(bus *bus.BusProvider, toolManager tool.ToolManagerInterface) *UpstreamWorker {
	return &UpstreamWorker{
		bus:         bus,
		toolManager: toolManager,
	}
}

// Start launches the worker in a new goroutine. It subscribes to tool execution
// requests on the event bus and will continue to process them until the
// provided context is canceled.
//
// ctx is the context that controls the lifecycle of the worker.
func (w *UpstreamWorker) Start(ctx context.Context) {
	log := logging.GetLogger().With("component", "UpstreamWorker")
	log.Info("Upstream worker started")

	requestBus := bus.GetBus[*bus.ToolExecutionRequest](w.bus, bus.ToolExecutionRequestTopic)
	resultBus := bus.GetBus[*bus.ToolExecutionResult](w.bus, bus.ToolExecutionResultTopic)

	unsubscribe := requestBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
		start := time.Now()
		metrics.IncrCounter([]string{"worker", "upstream", "request", "total"}, 1)
		defer metrics.MeasureSince([]string{"worker", "upstream", "request", "latency"}, start)
		log.Info("Received tool execution request", "tool", req.ToolName, "correlationID", req.CorrelationID())
		result, err := w.toolManager.CallTool(req.Context, &tool.ExecutionRequest{
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
		resultBus.Publish(ctx, req.CorrelationID(), res)
	})

	go func() {
		<-ctx.Done()
		log.Info("Upstream worker stopping")
		unsubscribe()
	}()
}
