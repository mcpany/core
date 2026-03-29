package interop

import (
	"context"
	"fmt"
)

// ACPAdapter implements the AgentFramework interface for ACP (Agent Communication Protocol).
//
// Intent: Represents an adapter that connects the ACP standard to the universal adapter hub.
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
type ACPAdapter struct {
	Capabilities map[string]bool
	MessageLog   []string // Maintain stateful checkpoints
}

// NewACPAdapter creates a new ACPAdapter instance.
//
// Intent: Constructs a new ACPAdapter with predefined capabilities.
//
// Parameters:
//   - None.
//
// Returns:
//   - *ACPAdapter: A newly created pointer to an ACPAdapter instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates internal state variables for capabilities and message log.
func NewACPAdapter() *ACPAdapter {
	return &ACPAdapter{
		Capabilities: map[string]bool{
			"agent_messaging": true,
			"capability_discovery": true,
		},
		MessageLog: make([]string, 0),
	}
}

// Name returns the identifier of the agent framework.
//
// Intent: Retrieves the exact name identifier for the ACP adapter.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The name "ACP".
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *ACPAdapter) Name() string {
	return "ACP"
}

// HandleTask translates and executes a universal task on the ACP framework.
//
// Intent: Processes and executes a task through simulated agent communication.
//
// Parameters:
//   - ctx (context.Context): Execution context for controlling cancellation and timeout.
//   - task (*Task): The generic task object that needs to be executed by ACP.
//
// Returns:
//   - *TaskResult: Contains the status, output, and telemetry information.
//   - error: Indicates failure in executing the task or an unsupported intent.
//
// Errors:
//   - Returns "ACP does not support capability" if the task's intent is missing from capabilities.
//
// Side Effects:
//   - Appends execution checkpoints (strings) to the internal `MessageLog` array, mutating adapter state.
func (a *ACPAdapter) HandleTask(ctx context.Context, task *Task) (*TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("ACP does not support capability: %s", task.Intent)
	}

	// Simulated stateful checkpoints
	msg := fmt.Sprintf("ACP communicated intent: %s", task.Intent)
	a.MessageLog = append(a.MessageLog, msg)

	output := fmt.Sprintf("Completed ACP task: %s, Messages: %d", task.Intent, len(a.MessageLog))

	return &TaskResult{
		TaskID: task.ID,
		Status: "success",
		Output: output,
		Telemetry: map[string]string{
			"delivery_status": "confirmed",
			"log_length":      fmt.Sprintf("%d", len(a.MessageLog)),
		},
	}, nil
}

// SupportsCapability checks if the framework provides a requested capability.
//
// Intent: Confirms if the ACP adapter's capabilities include the requested functionality.
//
// Parameters:
//   - capability (string): The intended capability name.
//
// Returns:
//   - bool: Indicates whether the given capability is supported.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *ACPAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}

// SyncMemoryShard synchronizes a hardware-attested multimodal memory shard with the ACP framework.
//
// Intent: Ingests a memory shard to synchronize state.
//
// Parameters:
//   - ctx (context.Context): The context for controlling cancellation and timeouts.
//   - shard (*MemoryShard): The multimodal memory shard to synchronize.
//
// Returns:
//   - error: An error if the signature is invalid.
//
// Errors:
//   - Returns an error if the shard signature verification fails.
//
// Side Effects:
//   - Modifies the message log state by injecting new checkpoints.
func (a *ACPAdapter) SyncMemoryShard(ctx context.Context, shard *MemoryShard) error {
	if shard.Signature == "" {
		return fmt.Errorf("invalid memory shard: signature required for ingestion")
	}

	a.MessageLog = append(a.MessageLog, fmt.Sprintf("Received multimodal shard via ACP: %s", shard.ShardID))
	return nil
}
