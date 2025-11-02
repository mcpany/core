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
	"sync"

	"al.essio.dev/pkg/shellescape"
	"github.com/alitto/pond/v2"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/logging"
)

// Worker is responsible for processing jobs from the bus.
type Worker struct {
	busProvider *bus.BusProvider
	pond        pond.Pool
	stopFuncs   []func()
	mu          sync.Mutex
}

// New creates a new Worker.
func New(busProvider *bus.BusProvider) *Worker {
	return &Worker{
		busProvider: busProvider,
		pond:        pond.NewPool(10, pond.WithQueueSize(100)),
	}
}

// Start starts the worker.
func (w *Worker) Start(ctx context.Context) {
	go w.startToolExecutionWorker(ctx)
}

// Stop stops the worker.
func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, stop := range w.stopFuncs {
		stop()
	}
	w.pond.StopAndWait()
}

func (w *Worker) startToolExecutionWorker(ctx context.Context) {
	reqBus := bus.GetBus[*bus.ToolExecutionRequest](w.busProvider, bus.ToolExecutionRequestTopic)
	resBus := bus.GetBus[*bus.ToolExecutionResult](w.busProvider, bus.ToolExecutionResultTopic)

	unsubscribe := reqBus.Subscribe(ctx, bus.ToolExecutionRequestTopic, func(req *bus.ToolExecutionRequest) {
		w.pond.Submit(func() {
			log := logging.GetLogger()
			log.Info("Received tool execution request", "tool_name", req.ToolName)
			// In a real implementation, this is where the tool would be
			// executed. For now, we'll just return a dummy result.
			result := shellescape.Quote(string(req.ToolInputs))
			res := &bus.ToolExecutionResult{
				BaseMessage: bus.BaseMessage{CID: req.CorrelationID()},
				Result:      []byte(result),
			}
			if err := resBus.Publish(context.Background(), req.CorrelationID(), res); err != nil {
				log.Error("Failed to publish tool execution result", "error", err)
			}
		})
	})
	w.mu.Lock()
	w.stopFuncs = append(w.stopFuncs, unsubscribe)
	w.mu.Unlock()
}
