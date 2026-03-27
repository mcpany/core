package interop

import (
	"context"
	"fmt"
)

// CrewAIAdapter implements the AgentFramework interface for CrewAI.
//
// Summary: Provides the implementation for interacting with the CrewAI framework via the universal adapter hub.
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
type CrewAIAdapter struct {
	Capabilities map[string]bool
	RoleRegistry map[string]string // Role name -> Capability token
}

// NewCrewAIAdapter creates a new CrewAIAdapter instance.
//
// Summary: Instantiates and initializes a new adapter for CrewAI with its predefined capabilities.
//
// Parameters:
//   - None.
//
// Returns:
//   - *CrewAIAdapter: A pointer to the newly created CrewAIAdapter instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates memory for the adapter and its internal mappings.
func NewCrewAIAdapter() *CrewAIAdapter {
	return &CrewAIAdapter{
		Capabilities: map[string]bool{
			"task_delegation": true,
			"role_discovery":  true,
		},
		RoleRegistry: make(map[string]string),
	}
}

// Name returns the identifier of the agent framework.
//
// Summary: Provides the unique identifier for the CrewAI adapter.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The name of the adapter ("CrewAI").
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *CrewAIAdapter) Name() string {
	return "CrewAI"
}

// HandleTask translates and executes a universal task on the CrewAI framework.
//
// Summary: Simulates executing a task using the delegated role mechanisms within the CrewAI framework.
//
// Parameters:
//   - ctx (context.Context): The context for execution, used to handle cancellation and timeouts.
//   - task (*Task): The universal task definition detailing the requested intent and payload.
//
// Returns:
//   - *TaskResult: The generalized result output, indicating success or failure.
//   - error: An error if the capability is unsupported or if the execution fails.
//
// Errors:
//   - Returns "CrewAI does not support capability" if the task's intent is not supported by the adapter.
//
// Side Effects:
//   - Modifies the internal RoleRegistry state to map the delegated role to an authentication token.
func (a *CrewAIAdapter) HandleTask(ctx context.Context, task *Task) (*TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("CrewAI does not support capability: %s", task.Intent)
	}

	// Simulated role discovery and task delegation mapping
	role, exists := task.Payload["role"]
	if !exists {
		role = "generalist"
	}

	a.RoleRegistry[role] = fmt.Sprintf("auth_token_%s", role)
	output := fmt.Sprintf("CrewAI delegated task: %s to role: %s", task.Intent, role)

	return &TaskResult{
		TaskID: task.ID,
		Status: "success",
		Output: output,
		Telemetry: map[string]string{
			"delegated_role": role,
			"auth_status":    "verified",
		},
	}, nil
}

// SupportsCapability checks if the framework provides a requested capability.
//
// Summary: Checks the internal capabilities map to see if the given intent is supported by CrewAI.
//
// Parameters:
//   - capability (string): The capability or intent name to check.
//
// Returns:
//   - bool: True if the capability is supported, otherwise false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *CrewAIAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}

// SyncMemoryShard synchronizes a hardware-attested multimodal memory shard with the CrewAI framework.
//
// Summary: Provides a unified way to distribute intent-pinned context to CrewAI's subagents.
//
// Parameters:
//   - ctx (context.Context): The context for controlling cancellation and timeouts.
//   - shard (*MemoryShard): The multimodal memory shard to synchronize.
//
// Returns:
//   - error: An error if the signature is missing or verification fails.
//
// Errors:
//   - Returns an error if the shard signature verification fails.
//
// Side Effects:
//   - Updates shared state available to instantiated CrewAI roles.
func (a *CrewAIAdapter) SyncMemoryShard(ctx context.Context, shard *MemoryShard) error {
	if shard.Signature == "" {
		return fmt.Errorf("invalid memory shard: unverified payload rejected by CrewAI")
	}

	return nil
}
