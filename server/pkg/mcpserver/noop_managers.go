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
// Summary: A tool manager that does nothing, useful for testing or placeholders.
type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Summary: No-op implementation of AddTool.
//
// Parameters:
//   - _ : tool.Tool. Unused.
//
// Returns:
//   - error: Always nil.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Summary: No-op implementation of GetTool.
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - tool.Tool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Summary: No-op implementation of ListTools.
//
// Returns:
//   - []tool.Tool: Always nil.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Summary: No-op implementation of ListMCPTools.
//
// Returns:
//   - []*mcp.Tool: Always nil.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Summary: No-op implementation of ClearToolsForService.
//
// Parameters:
//   - _ : string. Unused.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Summary: No-op implementation of ExecuteTool.
//
// Parameters:
//   - _ : context.Context. Unused.
//   - _ : *tool.ExecutionRequest. Unused.
//
// Returns:
//   - any: Always nil.
//   - error: Always nil.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Summary: No-op implementation of SetMCPServer.
//
// Parameters:
//   - _ : tool.MCPServerProvider. Unused.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Summary: No-op implementation of AddMiddleware.
//
// Parameters:
//   - _ : tool.ExecutionMiddleware. Unused.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: No-op implementation of AddServiceInfo.
//
// Parameters:
//   - _ : string. Unused.
//   - _ : *tool.ServiceInfo. Unused.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: No-op implementation of GetServiceInfo.
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - *tool.ServiceInfo: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Summary: No-op implementation of ListServices.
//
// Returns:
//   - []*tool.ServiceInfo: Always nil.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Summary: No-op implementation of SetProfiles.
//
// Parameters:
//   - _ : []string. Unused.
//   - _ : []*configv1.ProfileDefinition. Unused.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Summary: No-op implementation of IsServiceAllowed. Returns true to allow all in no-op.
//
// Parameters:
//   - _ : string. Unused.
//   - _ : string. Unused.
//
// Returns:
//   - bool: Always true.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Summary: No-op implementation of ToolMatchesProfile. Returns true.
//
// Parameters:
//   - _ : tool.Tool. Unused.
//   - _ : string. Unused.
//
// Returns:
//   - bool: Always true.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Summary: No-op implementation of GetAllowedServiceIDs.
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - map[string]bool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: No-op implementation of GetToolCountForService.
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - int: Always 0.
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// Summary: A prompt manager that does nothing.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Summary: No-op implementation of AddPrompt.
//
// Parameters:
//   - _ : prompt.Prompt. Unused.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Summary: No-op implementation of UpdatePrompt.
//
// Parameters:
//   - _ : prompt.Prompt. Unused.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Summary: No-op implementation of GetPrompt.
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - prompt.Prompt: Always nil.
//   - bool: Always false.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Summary: No-op implementation of ListPrompts.
//
// Returns:
//   - []prompt.Prompt: Always nil.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Summary: No-op implementation of ClearPromptsForService.
//
// Parameters:
//   - _ : string. Unused.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Summary: No-op implementation of SetMCPServer.
//
// Parameters:
//   - _ : prompt.MCPServerProvider. Unused.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary: A resource manager that does nothing.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Summary: No-op implementation of GetResource.
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - resource.Resource: Always nil.
//   - bool: Always false.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Summary: No-op implementation of AddResource.
//
// Parameters:
//   - _ : resource.Resource. Unused.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Summary: No-op implementation of RemoveResource.
//
// Parameters:
//   - _ : string. Unused.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Summary: No-op implementation of ListResources.
//
// Returns:
//   - []resource.Resource: Always nil.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Summary: No-op implementation of OnListChanged.
//
// Parameters:
//   - _ : func(). Unused.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Summary: No-op implementation of ClearResourcesForService.
//
// Parameters:
//   - _ : string. Unused.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
