package worker

import (
	"context"
	"sync"

	"al.essio.dev/pkg/shellescape"
	"github.com/alitto/pond/v2"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
)

// Config holds the configuration for the worker.
//
// Summary: Configuration for worker pool.
type Config struct {
	MaxWorkers   int
	MaxQueueSize int
}

// Worker is responsible for processing jobs from the bus.
//
// Summary: Processes background jobs.
type Worker struct {
	busProvider *bus.Provider
	pond        pond.Pool
	stopFuncs   []func()
	mu          sync.Mutex
	wg          sync.WaitGroup
}

// New creates a new Worker.
//
// Summary: Initializes a new Worker.
//
// Parameters:
//   - busProvider: *bus.Provider. The bus provider.
//   - cfg: *Config. The worker configuration.
//
// Returns:
//   - *Worker: The initialized worker.
func New(busProvider *bus.Provider, cfg *Config) *Worker {
	return &Worker{
		busProvider: busProvider,
		pond: pond.NewPool(
			cfg.MaxWorkers,
			pond.WithQueueSize(cfg.MaxQueueSize),
		),
	}
}

// Start starts the worker and its background tasks.
//
// Summary: Starts the worker processing loop.
//
// Parameters:
//   - ctx: context.Context. The context for the worker.
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.startToolExecutionWorker(ctx)
}

// Stop stops the worker and cleans up resources.
//
// Summary: Stops the worker.
//
// Side Effects:
//   - Waits for pending jobs.
//   - Unsubscribes from the bus.
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
