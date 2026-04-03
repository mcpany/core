// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"sync"

	"al.essio.dev/pkg/shellescape"
	"github.com/alitto/pond/v2"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
)

// Config represents the public Config entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Config struct {
	MaxWorkers   int
	MaxQueueSize int
}

// Worker represents the public Worker entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Worker struct {
	busProvider *bus.Provider
	pond        pond.Pool
	stopFuncs   []func()
	mu          sync.Mutex
	wg          sync.WaitGroup
}

// New serves as a public interface for interacting with New.
//
// Summary: Constructs and returns an initialized  ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func New(busProvider *bus.Provider, cfg *Config) *Worker {
	return &Worker{
		busProvider: busProvider,
		pond: pond.NewPool(
			cfg.MaxWorkers,
			pond.WithQueueSize(cfg.MaxQueueSize),
		),
	}
}

// Start serves as a public interface for interacting with Start.
//
// Summary: Start the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.startToolExecutionWorker(ctx)
}

// Stop serves as a public interface for interacting with Stop.
//
// Summary: Stop the  appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (w *Worker) Stop() {
	w.wg.Wait() // Wait for the subscription to be set up
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, stop := range w.stopFuncs {
		stop()
	}
	w.pond.StopAndWait()
}

func (w *Worker) startToolExecutionWorker(ctx context.Context) {
	defer w.wg.Done()
	log := logging.GetLogger()
	reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](w.busProvider, bus.ToolExecutionRequestTopic)
	if err != nil {
		log.Error("Failed to get request bus", "error", err)
		return
	}
	resBus, err := bus.GetBus[*bus.ToolExecutionResult](w.busProvider, bus.ToolExecutionResultTopic)
	if err != nil {
		log.Error("Failed to get result bus", "error", err)
		return
	}

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
	w.mu.Lock()
	w.stopFuncs = append(w.stopFuncs, unsubscribe)
	w.mu.Unlock()
}
