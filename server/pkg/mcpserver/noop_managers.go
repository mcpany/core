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
// Summary: is a no-op implementation of tool.ManagerInterface.
type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: tool.Tool. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - tool.Tool: The tool.Tool.
//   - bool: The bool.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   None.
//
// Returns:
//   - []tool.Tool: The []tool.Tool.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   None.
//
// Returns:
//   - []*mcp.Tool: The []*mcp.Tool.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   None.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: context.Context. The _.
//   - _: *tool.ExecutionRequest. The _.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: tool.MCPServerProvider. The _.
//
// Returns:
//   None.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: tool.ExecutionMiddleware. The _.
//
// Returns:
//   None.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//   - _: *tool.ServiceInfo. The _.
//
// Returns:
//   None.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - *tool.ServiceInfo: The *tool.ServiceInfo.
//   - bool: The bool.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   None.
//
// Returns:
//   - []*tool.ServiceInfo: The []*tool.ServiceInfo.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: []string. The _.
//   - _: []*configv1.ProfileDefinition. The _.
//
// Returns:
//   None.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//   - _: string. The _.
//
// Returns:
//   - bool: The bool.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: tool.Tool. The _.
//   - _: string. The _.
//
// Returns:
//   - bool: The bool.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - map[string]bool: The map[string]bool.
//   - bool: The bool.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - int: The int.
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// Summary: is a no-op implementation of prompt.ManagerInterface.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Summary: implements prompt.ManagerInterface.
//
// Parameters:
//   - _: prompt.Prompt. The _.
//
// Returns:
//   None.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Summary: implements prompt.ManagerInterface.
//
// Parameters:
//   - _: prompt.Prompt. The _.
//
// Returns:
//   None.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Summary: implements prompt.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - prompt.Prompt: The prompt.Prompt.
//   - bool: The bool.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Summary: implements prompt.ManagerInterface.
//
// Parameters:
//   None.
//
// Returns:
//   - []prompt.Prompt: The []prompt.Prompt.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Summary: implements prompt.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   None.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Summary: implements prompt.ManagerInterface.
//
// Parameters:
//   - _: prompt.MCPServerProvider. The _.
//
// Returns:
//   None.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary: is a no-op implementation of resource.ManagerInterface.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Summary: implements resource.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - resource.Resource: The resource.Resource.
//   - bool: The bool.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Summary: implements resource.ManagerInterface.
//
// Parameters:
//   - _: resource.Resource. The _.
//
// Returns:
//   None.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Summary: implements resource.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   None.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Summary: implements resource.ManagerInterface.
//
// Parameters:
//   None.
//
// Returns:
//   - []resource.Resource: The []resource.Resource.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Summary: implements resource.ManagerInterface.
//
// Parameters:
//   - _: func(). The _.
//
// Returns:
//   None.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Summary: implements resource.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   None.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
