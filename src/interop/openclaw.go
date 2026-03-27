package interop

import (
	"context"
	"fmt"
)

// OpenClawAdapter implements the AgentFramework interface for OpenClaw.
//
// Intent: An adapter implementation that bridges the OpenClaw agent framework with the universal hub.
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
type OpenClawAdapter struct {
	Capabilities map[string]bool
	CurrentEpoch int // Track the reasoning epoch
}

// NewOpenClawAdapter creates a new OpenClawAdapter instance.
//
// Intent: Initializes and returns a new adapter for OpenClaw with its specific capabilities.
//
// Params:
//   - None.
//
// Returns:
//   - *OpenClawAdapter: A pointer to the newly instantiated OpenClawAdapter.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates memory for the OpenClawAdapter and its capability map.
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
// Intent: Returns the specific name identifier of the OpenClaw adapter.
//
// Params:
//   - None.
//
// Returns:
//   - string: The name "OpenClaw".
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *OpenClawAdapter) Name() string {
	return "OpenClaw"
}

// HandleTask translates and executes a universal task on the OpenClaw framework.
//
// Intent: Executes the provided task using simulated adaptive reasoning logic.
//
// Params:
//   - ctx (context.Context): The task execution context, for managing lifecycle.
//   - task (*Task): The universal task object describing what to execute.
//
// Returns:
//   - *TaskResult: The generalized output from the executed task, along with telemetry data.
//   - error: An error indicating if the task failed or is unsupported.
//
// Errors:
//   - Returns an error if the framework's capability check fails for the task's intent.
//
// Side Effects:
//   - Increments the internal `CurrentEpoch` tracking state of the adapter.
func (a *OpenClawAdapter) HandleTask(ctx context.Context, task *Task) (*TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("OpenClaw does not support capability: %s", task.Intent)
	}

	// Simulated execution with state versioning logic (reasoning_epoch)
	a.CurrentEpoch++
	output := fmt.Sprintf("Executed OpenClaw task: %s, Epoch: %d", task.Intent, a.CurrentEpoch)

	return &TaskResult{
		TaskID: task.ID,
		Status: "success",
		Output: output,
		Telemetry: map[string]string{
			"reasoning_epoch": fmt.Sprintf("%d", a.CurrentEpoch),
			"entropy_score":   "low",
		},
	}, nil
}

// SupportsCapability checks if the framework provides a requested capability.
//
// Intent: Determines whether the OpenClaw adapter can execute tasks for a given capability intent.
//
// Params:
//   - capability (string): The capability identifier string to query.
//
// Returns:
//   - bool: True if the capability is found in the capabilities map; false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *OpenClawAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}
