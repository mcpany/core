package interop

import (
	"context"
	"fmt"
)

// AgentFramework represents a standardized AI framework to interact with.
//
// Summary: Defines the standard interface that all agent frameworks must implement to integrate with the universal bus.
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
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
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
	// Errors:
	//   - Returns an error if the framework does not support the requested capability or if execution fails.
	//
	// Side Effects:
	//   - May trigger external API calls or modify internal framework state depending on the task.
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
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
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
	// Errors:
	//   - Returns an error if the synchronization process fails.
	//
	// Side Effects:
	//   - Modifies internal framework state by updating the contextual memory.
	SyncMemoryShard(ctx context.Context, shard *MemoryShard) error
}

// MemoryShard represents a hardware-attested, intent-pinned memory fragment.
//
// Summary: A data structure that holds text context and an optional multimodal payload, with cryptographic lineage.
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
}

// Task represents a universal task definition for the Agent Bus.
//
// Summary: A data structure that holds the definition of a task to be routed to an agent framework.
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

// TaskResult holds the generalized output from an agent framework.
//
// Summary: A data structure that contains the result of an executed task.
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
}

// AdapterHub manages the registration and routing of tasks to different agent frameworks.
//
// Summary: A central hub that maintains registered agent adapters and routes tasks to the appropriate one.
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

// NewAdapterHub initializes a new AdapterHub.
//
// Summary: Creates and returns a new instance of AdapterHub with an empty registry.
//
// Parameters:
//   - None.
//
// Returns:
//   - *AdapterHub: A pointer to the newly created AdapterHub.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates memory for a new AdapterHub instance.
func NewAdapterHub() *AdapterHub {
	return &AdapterHub{
		adapters: make(map[string]AgentFramework),
	}
}

// RegisterAdapter adds a new framework adapter to the hub.
//
// Summary: Registers an agent framework adapter with the hub, allowing tasks to be routed to it.
//
// Parameters:
//   - adapter (AgentFramework): The adapter implementation to register.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the internal adapters map by adding a new entry.
func (h *AdapterHub) RegisterAdapter(adapter AgentFramework) {
	h.adapters[adapter.Name()] = adapter
}

// RouteTask finds the appropriate adapter for a task and executes it.
//
// Summary: Routes a given task to the registered framework adapter and returns the execution result.
//
// Parameters:
//   - ctx (context.Context): The context for controlling cancellation and timeouts.
//   - task (*Task): The universal task definition containing the framework and intent.
//
// Returns:
//   - *TaskResult: The result of the task execution.
//   - error: An error if routing or execution fails.
//
// Errors:
//   - Returns "no adapter registered for framework" if the requested framework is not found.
//   - Returns any error produced by the adapter during task execution.
//
// Side Effects:
//   - Triggers the task execution on the corresponding adapter, which may have its own side effects.
func (h *AdapterHub) RouteTask(ctx context.Context, task *Task) (*TaskResult, error) {
	adapter, exists := h.adapters[task.Framework]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for framework: %s", task.Framework)
	}
	return adapter.HandleTask(ctx, task)
}
