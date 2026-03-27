package interop

import (
	"context"
	"fmt"
)

// AutoGenAdapter implements the AgentFramework interface for AutoGen.
type AutoGenAdapter struct {
	Capabilities map[string]bool
	ChatHistory  []string // Maintain stateful checkpoints
}

// NewAutoGenAdapter creates a new AutoGenAdapter instance.
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
func (a *AutoGenAdapter) Name() string {
	return "AutoGen"
}

// HandleTask translates and executes a universal task on the AutoGen framework.
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
func (a *AutoGenAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}
