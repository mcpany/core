package interop

import (
	"context"
	"fmt"
)

// OpenClawAdapter implements the AgentFramework interface for OpenClaw.
//
// Summary: An adapter implementation that bridges the OpenClaw agent framework with the universal hub.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
//
// Side Effects:
//   - None.
type OpenClawAdapter struct {
	Capabilities map[string]bool
	CurrentEpoch int // Track the reasoning epoch
}

// NewOpenClawAdapter creates a new OpenClawAdapter instance.
//
// Summary: Initializes and returns a new adapter for OpenClaw with its specific capabilities.
//
// Parameters:
//   - None.
//
// Returns:
//   - *OpenClawAdapter: A pointer to the newly instantiated OpenClawAdapter.
//
// Throws/Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewOpenClawAdapter() *OpenClawAdapter {
	return &OpenClawAdapter{
		Capabilities: map[string]bool{
			"adaptive_reasoning": true,
			"context_sync":       true,
		},
		CurrentEpoch: 1,
	}
}

// Name returns the identifier of the agent framework.
//
// Summary: Returns the specific name identifier of the OpenClaw adapter.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The name "OpenClaw".
//
// Throws/Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *OpenClawAdapter) Name() string {
	return "OpenClaw"
}

// HandleTask translates and executes a universal task on the OpenClaw framework.
//
// Summary: Executes the provided task using simulated adaptive reasoning logic.
//
// Parameters:
//   - ctx (context.Context): The task execution context, for managing lifecycle.
//   - task (*Task): The universal task object describing what to execute.
//
// Returns:
//   - *TaskResult: The generalized output from the executed task, along with telemetry data.
//   - error: An error indicating if the task failed or is unsupported.
//
// Throws/Errors:
//   - Returns an error if the framework's capability check fails for the task's intent.
//
// Side Effects:
//   - None.
func (a *OpenClawAdapter) HandleTask(ctx context.Context, task *Task) (*TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("OpenClaw does not support capability: %s", task.Intent)
	}

	// Simulated execution with state versioning logic (reasoning_epoch)
	a.CurrentEpoch++
	output := fmt.Sprintf("Executed OpenClaw task: %s, Epoch: %d", task.Intent, a.CurrentEpoch)

	res := &TaskResult{
		TaskID: task.ID,
		Status: "success",
		Output: output,
		Telemetry: map[string]string{
			"reasoning_epoch": fmt.Sprintf("%d", a.CurrentEpoch),
			"entropy_score":   "low",
		},
	}

	if task.Payload["stream"] == "true" {
		res.Stream = make(chan string)
		go func() {
			res.Stream <- "chunk 1"
			res.Stream <- "chunk 2"
			close(res.Stream)
		}()
	}

	return res, nil
}

// SupportsCapability checks if the framework provides a requested capability.
//
// Summary: Determines whether the OpenClaw adapter can execute tasks for a given capability intent.
//
// Parameters:
//   - capability (string): The capability identifier string to query.
//
// Returns:
//   - bool: True if the capability is found in the capabilities map; false otherwise.
//
// Throws/Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *OpenClawAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}

// SyncMemoryShard synchronizes a hardware-attested multimodal memory shard with the OpenClaw framework.
//
// Summary: Ingests a memory shard and appends it to OpenClaw's internal state.
//
// Parameters:
//   - ctx (context.Context): The context for controlling cancellation and timeouts.
//   - shard (*MemoryShard): The multimodal memory shard to synchronize.
//
// Returns:
//   - error: An error if the signature is invalid.
//
// Throws/Errors:
//   - Returns an error if the shard signature verification fails.
//
// Side Effects:
//   - None.
func (a *OpenClawAdapter) SyncMemoryShard(ctx context.Context, shard *MemoryShard) error {
	if shard.Signature == "" {
		return fmt.Errorf("invalid memory shard: missing signature")
	}

	// Simulate processing the shard into OpenClaw's context.
	_ = shard.TextContent
	_ = shard.MultimodalPayload

	return nil
}

// StreamTask streams the execution of a task from the OpenClaw framework.
//
// Summary: Simulates a streaming task execution by emitting chunks to a channel.
//
// Parameters:
//   - ctx (context.Context): The context for execution, handling cancellation.
//   - task (*Task): The generic task object to execute.
//
// Returns:
//   - <-chan *TaskResult: A read-only channel emitting streamed chunks.
//   - error: Indicates failure in executing the task or an unsupported intent.
//
// Throws/Errors:
//   - Returns an error if the framework's capability check fails for the task's intent.
//
// Side Effects:
//   - Spawns a goroutine to send chunks.
func (a *OpenClawAdapter) StreamTask(ctx context.Context, task *Task) (<-chan *TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("OpenClaw does not support capability: %s", task.Intent)
	}

	stream := make(chan *TaskResult)

	go func() {
		defer close(stream)

		a.CurrentEpoch++

		// Send initial chunk
		select {
		case <-ctx.Done():
			return
		case stream <- &TaskResult{
			TaskID: task.ID,
			Status: "streaming",
			Output: fmt.Sprintf("Started OpenClaw task: %s, Epoch: %d", task.Intent, a.CurrentEpoch),
			Telemetry: map[string]string{
				"reasoning_epoch": fmt.Sprintf("%d", a.CurrentEpoch),
				"entropy_score":   "low",
				"chunk_index":     "0",
			},
		}:
		}

		// Send intermediate chunk
		select {
		case <-ctx.Done():
			return
		case stream <- &TaskResult{
			TaskID: task.ID,
			Status: "streaming",
			Output: fmt.Sprintf("Processing OpenClaw task: %s...", task.Intent),
			Telemetry: map[string]string{
				"reasoning_epoch": fmt.Sprintf("%d", a.CurrentEpoch),
				"entropy_score":   "low",
				"chunk_index":     "1",
			},
		}:
		}

		// Send final chunk
		select {
		case <-ctx.Done():
			return
		case stream <- &TaskResult{
			TaskID: task.ID,
			Status: "success",
			Output: fmt.Sprintf("Finished OpenClaw task: %s", task.Intent),
			Telemetry: map[string]string{
				"reasoning_epoch": fmt.Sprintf("%d", a.CurrentEpoch),
				"entropy_score":   "low",
				"chunk_index":     "2",
			},
		}:
		}
	}()

	return stream, nil
}
