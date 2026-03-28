package interop

import (
	"context"
	"fmt"
)

// ACPAdapter implements the AgentFramework interface for ACP.
//
// Intent: An adapter implementation that bridges the Agent Context Protocol (ACP) framework with the universal hub.
type ACPAdapter struct {
	Capabilities map[string]bool
	ContextSyncs int // Track the number of sync operations
}

// NewACPAdapter creates a new ACPAdapter instance.
//
// Intent: Initializes and returns a new adapter for ACP with its specific capabilities.
func NewACPAdapter() *ACPAdapter {
	return &ACPAdapter{
		Capabilities: map[string]bool{
			"acp_context_sync": true,
			"a2a_messaging":    true,
		},
		ContextSyncs: 0,
	}
}

// Name returns the identifier of the agent framework.
func (a *ACPAdapter) Name() string {
	return "ACP"
}

// HandleTask translates and executes a universal task on the ACP framework.
func (a *ACPAdapter) HandleTask(ctx context.Context, task *Task) (*TaskResult, error) {
	if !a.SupportsCapability(task.Intent) {
		return nil, fmt.Errorf("ACP does not support capability: %s", task.Intent)
	}

	a.ContextSyncs++
	output := fmt.Sprintf("Executed ACP task: %s, Syncs: %d", task.Intent, a.ContextSyncs)

	return &TaskResult{
		TaskID: task.ID,
		Status: "success",
		Output: output,
		Telemetry: map[string]string{
			"context_syncs": fmt.Sprintf("%d", a.ContextSyncs),
			"protocol":      "acp",
		},
	}, nil
}

// SupportsCapability checks if the framework provides a requested capability.
func (a *ACPAdapter) SupportsCapability(capability string) bool {
	return a.Capabilities[capability]
}

// SyncMemoryShard synchronizes a hardware-attested multimodal memory shard with the ACP framework.
func (a *ACPAdapter) SyncMemoryShard(ctx context.Context, shard *MemoryShard) error {
	if shard.Signature == "" {
		return fmt.Errorf("invalid memory shard: missing signature")
	}

	return nil
}
