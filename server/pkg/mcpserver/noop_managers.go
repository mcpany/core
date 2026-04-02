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
// AddTool implements tool.ManagerInterface.
// Summary: No-op AddTool.
// Parameters:
//   - _ (tool.Tool): Unused.
//
// Returns:
//   - error: Always returns nil.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// GetTool implements tool.ManagerInterface.
// Summary: No-op GetTool.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - tool.Tool: Always nil.
//   - bool: Always false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// ListTools implements tool.ManagerInterface.
// Summary: Returns an empty list of tools.
// Parameters:
//   - None.
//
// Returns:
//   - []tool.Tool: Always nil.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// ListMCPTools implements tool.ManagerInterface.
// Summary: Returns an empty list of MCP tools.
// Parameters:
//   - None.
//
// Returns:
//   - []*mcp.Tool: Always nil.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// ClearToolsForService implements tool.ManagerInterface.
// Summary: No-op ClearToolsForService.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// ExecuteTool implements tool.ManagerInterface.
// Summary: No-op ExecuteTool.
// Parameters:
//   - _ (context.Context): Unused.
//   - _ (*tool.ExecutionRequest): Unused.
//
// Returns:
//   - any: Always nil.
//   - error: Always nil.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// SetMCPServer implements tool.ManagerInterface.
// Summary: No-op SetMCPServer.
// Parameters:
//   - _ (tool.MCPServerProvider): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// AddMiddleware implements tool.ManagerInterface.
// Summary: No-op AddMiddleware.
// Parameters:
//   - _ (tool.ExecutionMiddleware): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// AddServiceInfo implements tool.ManagerInterface.
// Summary: No-op AddServiceInfo.
// Parameters:
//   - _ (string): Unused.
//   - _ (*tool.ServiceInfo): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// GetServiceInfo implements tool.ManagerInterface.
// Summary: No-op GetServiceInfo.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - *tool.ServiceInfo: Always nil.
//   - bool: Always false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// ListServices implements tool.ManagerInterface.
// Summary: Returns an empty list of services.
// Parameters:
//   - None.
//
// Returns:
//   - []*tool.ServiceInfo: Always nil.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// SetProfiles implements tool.ManagerInterface.
// Summary: No-op SetProfiles.
// Parameters:
//   - _ ([]string): Unused.
//   - _ ([]*configv1.ProfileDefinition): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// IsServiceAllowed implements tool.ManagerInterface.
// Summary: No-op IsServiceAllowed.
// Parameters:
//   - _, _ (string): Unused.
//
// Returns:
//   - bool: Always true (allow all).
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// ToolMatchesProfile implements tool.ManagerInterface.
// Summary: No-op ToolMatchesProfile.
// Parameters:
//   - _ (tool.Tool): Unused.
//   - _ (string): Unused.
//
// Returns:
//   - bool: Always true.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// GetAllowedServiceIDs implements tool.ManagerInterface.
// Summary: No-op GetAllowedServiceIDs.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - map[string]bool: Always nil.
//   - bool: Always false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// GetToolCountForService implements tool.ManagerInterface.
// Summary: No-op GetToolCountForService.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - int: Always 0.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type NoOpToolManager struct{}

func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

func (m *NoOpToolManager) ClearToolsForService(_ string) {}

func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// Summary: A prompt manager that does nothing.
// AddPrompt implements prompt.ManagerInterface.
// Summary: No-op AddPrompt.
// Parameters:
//   - _ (prompt.Prompt): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// UpdatePrompt implements prompt.ManagerInterface.
// Summary: No-op UpdatePrompt.
// Parameters:
//   - _ (prompt.Prompt): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// GetPrompt implements prompt.ManagerInterface.
// Summary: No-op GetPrompt.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - prompt.Prompt: Always nil.
//   - bool: Always false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// ListPrompts implements prompt.ManagerInterface.
// Summary: Returns an empty list of prompts.
// Parameters:
//   - None.
//
// Returns:
//   - []prompt.Prompt: Always nil.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// ClearPromptsForService implements prompt.ManagerInterface.
// Summary: No-op ClearPromptsForService.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// SetMCPServer implements prompt.ManagerInterface.
// Summary: No-op SetMCPServer.
// Parameters:
//   - _ (prompt.MCPServerProvider): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
type NoOpPromptManager struct{}

func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary: A resource manager that does nothing.
// GetResource implements resource.ManagerInterface.
// Summary: No-op GetResource.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - resource.Resource: Always nil.
//   - bool: Always false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// AddResource implements resource.ManagerInterface.
// Summary: No-op AddResource.
// Parameters:
//   - _ (resource.Resource): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// RemoveResource implements resource.ManagerInterface.
// Summary: No-op RemoveResource.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// ListResources implements resource.ManagerInterface.
// Summary: Returns an empty list of resources.
// Parameters:
//   - None.
//
// Returns:
//   - []resource.Resource: Always nil.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// OnListChanged implements resource.ManagerInterface.
// Summary: No-op OnListChanged.
// Parameters:
//   - _ (func()): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
//
// ClearResourcesForService implements resource.ManagerInterface.
// Summary: No-op ClearResourcesForService.
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors:
//   - None.
type NoOpResourceManager struct{}

func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

func (m *NoOpResourceManager) RemoveResource(_ string) {}

func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

func (m *NoOpResourceManager) OnListChanged(_ func()) {}

func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
