/*
 * Copyright 2025 Author(s) of MCP-XY
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

	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/tool"
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

	requestBus := bus.GetBus[*bus.ToolExecutionRequest](w.bus, "tool_execution_requests")
	resultBus := bus.GetBus[*bus.ToolExecutionResult](w.bus, "tool_execution_results")

	unsubscribe := requestBus.Subscribe("request", func(req *bus.ToolExecutionRequest) {
		log.Info("Received tool execution request", "tool", req.ToolName, "correlationID", req.CorrelationID())
		result, err := w.toolManager.ExecuteTool(req.Context, &tool.ExecutionRequest{
			ToolName:   req.ToolName,
			ToolInputs: req.ToolInputs,
		})

		var resultBytes json.RawMessage
		if err == nil {
			resultBytes, err = json.Marshal(result)
		}

		res := &bus.ToolExecutionResult{
			Result: resultBytes,
			Error:  err,
		}
		res.SetCorrelationID(req.CorrelationID())
		resultBus.Publish(req.CorrelationID(), res)
	})

	go func() {
		<-ctx.Done()
		log.Info("Upstream worker stopping")
		unsubscribe()
	}()
}
