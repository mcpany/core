package interop

import (
	"context"
	"fmt"
)

// AgentFramework represents the public AgentFramework entity.
//
// Summary: Defines the structured data model representing a framework.
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
type AgentFramework interface {
	// Name returns the identifier of the agent framework.
	//
	// Summary: Retrieves the unique name of the framework.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - string: The name of the agent framework.
	//
	// Throws/Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	//
	Name() string

	// HandleTask receives a task from the universal bus, translates it, executes it, and returns the result.
	//
	// Summary: Translates and executes a given task within the specific agent framework.
	//
	// Parameters:
	//   - ctx (context.Context): The context for controlling cancellation and timeouts.
	//   - task (*Task): The universal task definition to execute.
	//
	// Returns:
	//   - *TaskResult: The generalized output from the agent framework.
	//   - error: An error if the task execution fails.
	//
	// Throws/Errors:
	//   - Returns an error if the framework does not support the requested capability or if execution fails.
	//
	// Side Effects:
	//   - None.
	//
	HandleTask(ctx context.Context, task *Task) (*TaskResult, error)

	// SupportsCapability checks if the framework provides a requested capability.
	//
	// Summary: Verifies whether the framework can handle a specific intent or capability.
	//
	// Parameters:
	//   - capability (string): The capability or intent to check.
	//
	// Returns:
	//   - bool: True if the capability is supported, false otherwise.
	//
	// Throws/Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	//
	SupportsCapability(capability string) bool

	// SyncMemoryShard synchronizes a hardware-attested multimodal memory shard with the agent framework.
	//
	// Summary: Syncs a state shard with the framework, ensuring multimodal trace integrity.
	//
	// Parameters:
	//   - ctx (context.Context): The context for controlling cancellation and timeouts.
	//   - shard (*MemoryShard): The multimodal memory shard to synchronize.
	//
	// Returns:
	//   - error: An error if the shard synchronization fails or signature is invalid.
	//
	// Throws/Errors:
	//   - Returns an error if the synchronization process fails.
	//
	// Side Effects:
	//   - None.
	//
	SyncMemoryShard(ctx context.Context, shard *MemoryShard) error

	// StreamTask receives a task from the universal bus and streams the result back in chunks.
	//
	// Summary: Executes a task within the framework and emits streaming results as they become available.
	//
	// Parameters:
	//   - ctx (context.Context): The context for controlling cancellation and timeouts.
	//   - task (*Task): The universal task definition to execute.
	//
	// Returns:
	//   - <-chan *TaskResult: A read-only channel emitting streamed task results.
	//   - error: An error if the task execution fails to start.
	//
	// Throws/Errors:
	//   - Returns an error if the framework does not support the capability or initialization fails.
	//
	// Side Effects:
	//   - Spawns a goroutine to stream results back through the channel.
	StreamTask(ctx context.Context, task *Task) (<-chan *TaskResult, error)
}

// MemoryShard represents the public MemoryShard entity.
//
// Summary: Defines the structured data model representing a shard.
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
type MemoryShard struct {
	ShardID           string `json:"shard_id"`
	Intent            string `json:"intent"`
	TextContent       string `json:"text_content"`
	MultimodalPayload []byte `json:"multimodal_payload,omitempty"`
	Signature         string `json:"signature"`
	PreviousHash      string `json:"previous_hash,omitempty"` // For Multimodal Hash-Chaining (MHC)
}

// Task represents the public Task entity.
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
type Task struct {
	ID        string            `json:"id"`
	Framework string            `json:"framework"`
	Intent    string            `json:"intent"`
	Payload   map[string]string `json:"payload"`
}

// TaskResult represents the public TaskResult entity.
//
// Summary: Defines the structured data model representing a result.
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
type TaskResult struct {
	TaskID    string            `json:"task_id"`
	Status    string            `json:"status"`
	Output    string            `json:"output"`
	Telemetry map[string]string `json:"telemetry,omitempty"`
	Stream    chan string       `json:"-"`
}

// AdapterHub represents the public AdapterHub entity.
//
// Summary: Defines the structured data model representing a hub.
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
type AdapterHub struct {
	adapters map[string]AgentFramework
}

// NewAdapterHub serves as a public interface for interacting with NewAdapterHub.
//
// Summary: Constructs and returns an initialized adapter hub ready for consumption.
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
func NewAdapterHub() *AdapterHub {
	return &AdapterHub{
		adapters: make(map[string]AgentFramework),
	}
}

// RegisterAdapter serves as a public interface for interacting with RegisterAdapter.
//
// Summary: Register the adapter appropriately based on current system conditions.
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
func (h *AdapterHub) RegisterAdapter(adapter AgentFramework) {
	h.adapters[adapter.Name()] = adapter
}

// RouteTask serves as a public interface for interacting with RouteTask.
//
// Summary: Route the task appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *AdapterHub) RouteTask(ctx context.Context, task *Task) (*TaskResult, error) {
	adapter, exists := h.adapters[task.Framework]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for framework: %s", task.Framework)
	}
	return adapter.HandleTask(ctx, task)
}

// StreamRouteTask serves as a public interface for interacting with StreamRouteTask.
//
// Summary: Stream the route task appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *AdapterHub) StreamRouteTask(ctx context.Context, task *Task) (<-chan *TaskResult, error) {
	adapter, exists := h.adapters[task.Framework]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for framework: %s", task.Framework)
	}
	return adapter.StreamTask(ctx, task)
}
