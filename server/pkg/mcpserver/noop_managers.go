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
// Summary: A tool manager that does nothing.
type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Summary: No-op AddTool.
//
// Parameters: - None.
//   - _ (tool.Tool): Unused.
//
// Returns: - None.
//   - error: Always returns nil.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Summary: No-op GetTool.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - tool.Tool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Summary: Returns an empty list of tools.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []tool.Tool: Always nil.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Summary: Returns an empty list of MCP tools.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []*mcp.Tool: Always nil.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Summary: No-op ClearToolsForService.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Summary: No-op ExecuteTool.
//
// Parameters: - None.
//   - _ (context.Context): Unused.
//   - _ (*tool.ExecutionRequest): Unused.
//
// Returns: - None.
//   - any: Always nil.
//   - error: Always nil.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Summary: No-op SetMCPServer.
//
// Parameters: - None.
//   - _ (tool.MCPServerProvider): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Summary: No-op AddMiddleware.
//
// Parameters: - None.
//   - _ (tool.ExecutionMiddleware): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: No-op AddServiceInfo.
//
// Parameters: - None.
//   - _ (string): Unused.
//   - _ (*tool.ServiceInfo): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: No-op GetServiceInfo.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - *tool.ServiceInfo: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Summary: Returns an empty list of services.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []*tool.ServiceInfo: Always nil.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Summary: No-op SetProfiles.
//
// Parameters: - None.
//   - _ ([]string): Unused.
//   - _ ([]*configv1.ProfileDefinition): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Summary: No-op IsServiceAllowed.
//
// Parameters: - None.
//   - _, _ (string): Unused.
//
// Returns: - None.
//   - bool: Always true (allow all).
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Summary: No-op ToolMatchesProfile.
//
// Parameters: - None.
//   - _ (tool.Tool): Unused.
//   - _ (string): Unused.
//
// Returns: - None.
//   - bool: Always true.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Summary: No-op GetAllowedServiceIDs.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - map[string]bool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: No-op GetToolCountForService.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
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
// Summary: No-op AddPrompt.
//
// Parameters: - None.
//   - _ (prompt.Prompt): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Summary: No-op UpdatePrompt.
//
// Parameters: - None.
//   - _ (prompt.Prompt): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Summary: No-op GetPrompt.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - prompt.Prompt: Always nil.
//   - bool: Always false.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Summary: Returns an empty list of prompts.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []prompt.Prompt: Always nil.
//
// Side Effects: - None.
//   - None.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Summary: No-op ClearPromptsForService.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Summary: No-op SetMCPServer.
//
// Parameters: - None.
//   - _ (prompt.MCPServerProvider): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary: A resource manager that does nothing.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Summary: No-op GetResource.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - resource.Resource: Always nil.
//   - bool: Always false.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Summary: No-op AddResource.
//
// Parameters: - None.
//   - _ (resource.Resource): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Summary: No-op RemoveResource.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Summary: Returns an empty list of resources.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []resource.Resource: Always nil.
//
// Side Effects: - None.
//   - None.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Summary: No-op OnListChanged.
//
// Parameters: - None.
//   - _ (func()): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Summary: No-op ClearResourcesForService.
//
// Parameters: - None.
//   - _ (string): Unused.
//
// Returns: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
