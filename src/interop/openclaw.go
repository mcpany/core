package interop

import (
	"context"
	"fmt"
)

// OpenClawAdapter implements the AgentFramework interface for OpenClaw.
type OpenClawAdapter struct {
	Capabilities map[string]bool
	CurrentEpoch int // Track the reasoning epoch
}

// NewOpenClawAdapter creates a new OpenClawAdapter instance.
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
func (a *OpenClawAdapter) Name() string {
	return "OpenClaw"
}

// HandleTask translates and executes a universal task on the OpenClaw framework.
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
func (a *OpenClawAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}
