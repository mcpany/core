package interop

import (
	"context"
	"fmt"
)

// AgentFramework represents a standardized AI framework to interact with.
//
// Summary: Defines the standard interface that all agent frameworks must implement to integrate with the universal bus.
//
//
// Parameters:
//   - Not applicable for a type.
//
//
// Returns:
//   - Not applicable for a type.
//
//
// Errors:
//   - Not applicable for a type.
//
//
// Side Effects:
//   - Implementing this interface allows a framework to be registered with the universal bus.
type AgentFramework interface {
	// Name returns the identifier of the agent framework.
	//
	// Summary: Retrieves the unique name of the framework.
	//
	// Parameters:
	//   - This function does not accept any parameters.
	//
	// Returns:
	//   - string: The statically defined name of the agent framework.
	//
	// Errors:
	//   - Does not produce any errors.
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
	//   - *TaskResult: A populated TaskResult struct containing the framework's output.
	//   - error: An error object indicating failure reasons during task handling.
	//
	// Errors:
	//   - Throws an error if the task payload is malformed or the execution strategy fails.
	//
	HandleTask(ctx context.Context, task *Task) (*TaskResult, error)

	// SupportsCapability checks if the framework provides a requested capability.
	//
	// Summary: Verifies whether the framework can handle a specific intent or capability.
	//
	// Parameters:
	//   - capability (string): The capability identifier or intent string to evaluate.
	//
	// Returns:
	//   - bool: True if the adapter has registered support for the capability; false otherwise.
	//
	// Errors:
	//   - This function does not produce any errors.
	//
	SupportsCapability(capability string) bool

	// SyncMemoryShard synchronizes a hardware-attested multimodal memory shard with the agent framework.
	//
	// Summary: Syncs a state shard with the framework, ensuring multimodal trace integrity.
	//
	// Parameters:
	//   - ctx (context.Context): The context for controlling cancellation and timeouts.
	//   - shard (*MemoryShard): The multimodal memory shard to synchronize, containing text or binary payloads.
	//
	// Returns:
	//   - error: An error indicating synchronization failure, typically due to an invalid cryptographic signature.
	//
	// Errors:
	//   - Throws an error if the memory shard cannot be verified or successfully ingested into the framework's state.
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
	// Errors:
	//   - Returns an error if the framework does not support the capability or initialization fails.
	//
	// Side Effects:
	//   - Implementation dependent. Spawns a goroutine to stream results back through the channel.
	StreamTask(ctx context.Context, task *Task) (<-chan *TaskResult, error)
}

// MemoryShard represents a hardware-attested, intent-pinned memory fragment.
//
// Summary: A data structure that holds text context and an optional multimodal payload, with cryptographic lineage.
//
//
// Parameters:
//   - ShardID (string): A unique identifier for the memory shard.
//   - Intent (string): The capability intent pinned to this memory state.
//   - TextContent (string): The raw text or prompt state.
//   - MultimodalPayload ([]byte): Optional binary payload for multimodal state.
//   - Signature (string): Cryptographic signature proving shard provenance.
//   - PreviousHash (string): Hash of the preceding shard for chaining.
//
//
// Returns:
//   - Not applicable for a type.
//
//
// Errors:
//   - Not applicable for a type.
//
//
// Side Effects:
//   - Instantiating this type has no side effects.
type MemoryShard struct {
	ShardID           string `json:"shard_id"`
	Intent            string `json:"intent"`
	TextContent       string `json:"text_content"`
	MultimodalPayload []byte `json:"multimodal_payload,omitempty"`
	Signature         string `json:"signature"`
	PreviousHash      string `json:"previous_hash,omitempty"` // For Multimodal Hash-Chaining (MHC)
}

// Task represents a universal task definition for the Agent Bus.
//
// Summary: A data structure that holds the definition of a task to be routed to an agent framework.
//
//
// Parameters:
//   - ID (string): Unique tracking identifier for the task.
//   - Framework (string): The target agent framework to route to.
//   - Intent (string): The capability required to resolve the task.
//   - Payload (map[string]string): Additional dynamic arguments for execution.
//
//
// Returns:
//   - Not applicable for a type.
//
//
// Errors:
//   - Not applicable for a type.
//
//
// Side Effects:
//   - Instantiating this type has no side effects.
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
//
// Parameters:
//   - TaskID (string): Corresponds to the Task.ID being processed.
//   - Status (string): The outcome status (e.g., success, failed, running).
//   - Output (string): The resulting text from the agent framework.
//   - Telemetry (map[string]string): Execution metrics such as latency.
//   - Stream (chan string): Channel for streaming intermediate results.
//
//
// Returns:
//   - Not applicable for a type.
//
//
// Errors:
//   - Not applicable for a type.
//
//
// Side Effects:
//   - Instantiating this type has no side effects.
type TaskResult struct {
	TaskID    string            `json:"task_id"`
	Status    string            `json:"status"`
	Output    string            `json:"output"`
	Telemetry map[string]string `json:"telemetry,omitempty"`
	Stream    chan string       `json:"-"`
}

// AdapterHub manages the registration and routing of tasks to different agent frameworks.
//
// Summary: A central hub that maintains registered agent adapters and routes tasks to the appropriate one.
//
//
// Parameters:
//   - adapters (map[string]AgentFramework): Internal registry mapping framework names to their adapter implementations.
//
//
// Returns:
//   - Not applicable for a type.
//
//
// Errors:
//   - Not applicable for a type.
//
//
// Side Effects:
//   - Maintains state for registered capabilities across the universal bus.
type AdapterHub struct {
	adapters map[string]AgentFramework
}

// NewAdapterHub initializes a new AdapterHub.
//
// Summary: Creates and returns a new instance of AdapterHub with an empty registry.
//
//
// Parameters:
//   - This function does not accept any parameters.
//
//
// Returns:
//   - *AdapterHub: A pointer to the newly created AdapterHub instance, pre-allocated with an empty adapter map.
//
//
// Errors:
//   - This function does not produce any errors.
//
//
// Side Effects:
//   - Allocates memory for the AdapterHub map.
func NewAdapterHub() *AdapterHub {
	return &AdapterHub{
		adapters: make(map[string]AgentFramework),
	}
}

// RegisterAdapter adds a new framework adapter to the hub.
//
// Summary: Registers an agent framework adapter with the hub, allowing tasks to be routed to it.
//
//
// Parameters:
//   - adapter (AgentFramework): The adapter instance that handles the framework's capabilities.
//
//
// Returns:
//   - This function does not return any value.
//
//
// Errors:
//   - This function does not produce any errors.
//
//
// Side Effects:
//   - Modifies the internal adapters map of the AdapterHub by adding or updating an entry.
func (h *AdapterHub) RegisterAdapter(adapter AgentFramework) {
	h.adapters[adapter.Name()] = adapter
}

// RouteTask finds the appropriate adapter for a task and executes it.
//
// Summary: Routes a given task to the registered framework adapter and returns the execution result.
//
//
// Parameters:
//   - ctx (context.Context): The context for controlling cancellation and timeouts.
//   - task (*Task): The universal task definition containing the framework and intent.
//
//
// Returns:
//   - *TaskResult: The result of the task execution.
//   - error: An error if routing or execution fails.
//
//
// Errors:
//   - Returns "no adapter registered for framework" if the requested framework is not found.
//   - Returns any error produced by the adapter during task execution.
//
//
// Side Effects:
//   - This function has no side effects.
func (h *AdapterHub) RouteTask(ctx context.Context, task *Task) (*TaskResult, error) {
	adapter, exists := h.adapters[task.Framework]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for framework: %s", task.Framework)
	}
	return adapter.HandleTask(ctx, task)
}

// StreamRouteTask finds the appropriate adapter for a task and streams its execution.
//
// Summary: Routes a given task to the registered framework adapter and streams the execution result.
//
//
// Parameters:
//   - ctx (context.Context): The context for controlling cancellation and timeouts.
//   - task (*Task): The universal task definition containing the framework and intent.
//
//
// Returns:
//   - <-chan *TaskResult: A read-only channel emitting streamed task results.
//   - error: An error if routing or execution fails to start.
//
//
// Errors:
//   - Returns "no adapter registered for framework" if the requested framework is not found.
//   - Returns any error produced by the adapter during task execution initialization.
//
//
// Side Effects:
//   - Triggers the streaming task execution on the corresponding adapter.
func (h *AdapterHub) StreamRouteTask(ctx context.Context, task *Task) (<-chan *TaskResult, error) {
	adapter, exists := h.adapters[task.Framework]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for framework: %s", task.Framework)
	}
	return adapter.StreamTask(ctx, task)
}
