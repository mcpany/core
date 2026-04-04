// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NoOpToolManager is a no-op implementation of tool.ManagerInterface.
//
// Summary: Is a no-op implementation of tool.ManagerInterface.
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

type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (tool.Tool): Parameter.
//
// Returns:
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - tool.Tool: Return value.
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []tool.Tool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*mcp.Tool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (context.Context): Parameter.
//   - unnamed (*tool.ExecutionRequest): Parameter.
//
// Returns:
//   - any: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (tool.MCPServerProvider): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (tool.ExecutionMiddleware): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//   - unnamed (*tool.ServiceInfo): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - *tool.ServiceInfo: Return value.
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*tool.ServiceInfo: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed ([]string): Parameter.
//   - unnamed ([]*configv1.ProfileDefinition): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//   - unnamed (string): Parameter.
//
// Returns:
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (tool.Tool): Parameter.
//   - unnamed (string): Parameter.
//
// Returns:
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - map[string]bool: Return value.
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: Implements tool.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - int: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// Summary: Is a no-op implementation of prompt.ManagerInterface.
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

type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Summary: Implements prompt.ManagerInterface.
//
// Parameters:
//   - unnamed (prompt.Prompt): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Summary: Implements prompt.ManagerInterface.
//
// Parameters:
//   - unnamed (prompt.Prompt): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Summary: Implements prompt.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - prompt.Prompt: Return value.
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Summary: Implements prompt.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []prompt.Prompt: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Summary: Implements prompt.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Summary: Implements prompt.ManagerInterface.
//
// Parameters:
//   - unnamed (prompt.MCPServerProvider): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary: Is a no-op implementation of resource.ManagerInterface.
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

type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Summary: Implements resource.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - resource.Resource: Return value.
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Summary: Implements resource.ManagerInterface.
//
// Parameters:
//   - unnamed (resource.Resource): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Summary: Implements resource.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Summary: Implements resource.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []resource.Resource: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Summary: Implements resource.ManagerInterface.
//
// Parameters:
//   - unnamed (func()): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Summary: Implements resource.ManagerInterface.
//
// Parameters:
//   - unnamed (string): Parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
