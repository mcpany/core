package interop

import (
	"context"
	"fmt"
)

// AgentFramework represents a standardized AI framework to interact with.
type AgentFramework interface {
	// Name returns the identifier of the agent framework.
	Name() string
	// HandleTask receives a task from the universal bus, translates it, executes it, and returns the result.
	HandleTask(ctx context.Context, task *Task) (*TaskResult, error)
	// SupportsCapability checks if the framework provides a requested capability.
	SupportsCapability(capability string) bool
}

// Task represents a universal task definition for the Agent Bus.
type Task struct {
	ID        string            `json:"id"`
	Framework string            `json:"framework"`
	Intent    string            `json:"intent"`
	Payload   map[string]string `json:"payload"`
}

// TaskResult holds the generalized output from an agent framework.
type TaskResult struct {
	TaskID    string            `json:"task_id"`
	Status    string            `json:"status"`
	Output    string            `json:"output"`
	Telemetry map[string]string `json:"telemetry,omitempty"`
}

// AdapterHub manages the registration and routing of tasks to different agent frameworks.
type AdapterHub struct {
	adapters map[string]AgentFramework
}

// NewAdapterHub initializes a new AdapterHub.
func NewAdapterHub() *AdapterHub {
	return &AdapterHub{
		adapters: make(map[string]AgentFramework),
	}
}

// RegisterAdapter adds a new framework adapter to the hub.
func (h *AdapterHub) RegisterAdapter(adapter AgentFramework) {
	h.adapters[adapter.Name()] = adapter
}

// RouteTask finds the appropriate adapter for a task and executes it.
func (h *AdapterHub) RouteTask(ctx context.Context, task *Task) (*TaskResult, error) {
	adapter, exists := h.adapters[task.Framework]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for framework: %s", task.Framework)
	}
	return adapter.HandleTask(ctx, task)
}
