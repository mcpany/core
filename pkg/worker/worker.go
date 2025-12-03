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

// Config holds the configuration for the worker.
type Config struct {
	MaxWorkers   int
	MaxQueueSize int
}

// Worker is responsible for processing jobs from the bus.
type Worker struct {
	busProvider *bus.BusProvider
	pond        pond.Pool
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// New creates a new Worker.
func New(busProvider *bus.BusProvider, cfg *Config) *Worker {
	return &Worker{
		busProvider: busProvider,
		pond: pond.NewPool(
			cfg.MaxWorkers,
			pond.WithQueueSize(cfg.MaxQueueSize),
		),
	}
}

// Start starts the worker.
func (w *Worker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	startWg := &sync.WaitGroup{}
	startWg.Add(1)
	w.wg.Add(1)
	go w.startToolExecutionWorker(workerCtx, startWg)
	startWg.Wait() // wait for subscribe to happen
}

// Stop stops the worker.
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait() // Wait for the goroutine to finish
	w.pond.StopAndWait()
}

func (w *Worker) startToolExecutionWorker(ctx context.Context, startWg *sync.WaitGroup) {
	defer w.wg.Done()
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
			if err := resBus.Publish(ctx, req.CorrelationID(), res); err != nil {
				log.Error("Failed to publish tool execution result", "error", err)
			}
		})
	})
	defer unsubscribe()
	startWg.Done()

	<-ctx.Done()
}
