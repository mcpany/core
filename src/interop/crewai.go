package interop

import (
	"context"
	"fmt"
)

// CrewAIAdapter implements the AgentFramework interface for CrewAI.
type CrewAIAdapter struct {
	Capabilities map[string]bool
	RoleRegistry map[string]string // Role name -> Capability token
}

// NewCrewAIAdapter creates a new CrewAIAdapter instance.
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
func (a *CrewAIAdapter) Name() string {
	return "CrewAI"
}

// HandleTask translates and executes a universal task on the CrewAI framework.
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
func (a *CrewAIAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}
