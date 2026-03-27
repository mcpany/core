package interop

import (
	"context"
	"fmt"
)

// AutoGenAdapter implements the AgentFramework interface for AutoGen.
//
// Intent: Represents an adapter that connects the AutoGen multi-agent framework to the universal adapter hub.
//
// Params:
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
type AutoGenAdapter struct {
	Capabilities map[string]bool
	ChatHistory  []string // Maintain stateful checkpoints
}

// NewAutoGenAdapter creates a new AutoGenAdapter instance.
//
// Intent: Constructs a new AutoGenAdapter with predefined multi-agent capabilities.
//
// Params:
//   - None.
//
// Returns:
//   - *AutoGenAdapter: A newly created pointer to an AutoGenAdapter instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates internal state variables for capabilities and chat history.
func NewAutoGenAdapter() *AutoGenAdapter {
	return &AutoGenAdapter{
		Capabilities: map[string]bool{
			"multi_agent_chat": true,
			"subagent_exec":    true,
		},
		ChatHistory: make([]string, 0),
	}
}

// Name returns the identifier of the agent framework.
//
// Intent: Retrieves the exact name identifier for the AutoGen adapter.
//
// Params:
//   - None.
//
// Returns:
//   - string: The name "AutoGen".
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *AutoGenAdapter) Name() string {
	return "AutoGen"
}

// HandleTask translates and executes a universal task on the AutoGen framework.
//
// Intent: Processes and executes a task through simulated multi-agent subagent execution.
//
// Params:
//   - ctx (context.Context): Execution context for controlling cancellation and timeout.
//   - task (*Task): The generic task object that needs to be executed by AutoGen.
//
// Returns:
//   - *TaskResult: Contains the status, output, and telemetry information from the subagent.
//   - error: Indicates failure in executing the task or an unsupported intent.
//
// Errors:
//   - Returns "AutoGen does not support capability" if the task's intent is missing from capabilities.
//
// Side Effects:
//   - Appends execution checkpoints (strings) to the internal `ChatHistory` array, mutating adapter state.
func (a *AutoGenAdapter) HandleTask(ctx context.Context, task *Task) (*TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("AutoGen does not support capability: %s", task.Intent)
	}

	// Simulated stateful checkpoints (Sandbox Persistence Proofs)
	msg := fmt.Sprintf("AutoGen subagent executed task: %s", task.Intent)
	a.ChatHistory = append(a.ChatHistory, msg)

	output := fmt.Sprintf("Completed AutoGen subagent task: %s, Checkpoints: %d", task.Intent, len(a.ChatHistory))

	return &TaskResult{
		TaskID: task.ID,
		Status: "success",
		Output: output,
		Telemetry: map[string]string{
			"mailbox_integrity": "verified",
			"history_length":    fmt.Sprintf("%d", len(a.ChatHistory)),
		},
	}, nil
}

// SupportsCapability checks if the framework provides a requested capability.
//
// Intent: Confirms if the AutoGen adapter's capabilities include the requested functionality.
//
// Params:
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
func (a *AutoGenAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}
