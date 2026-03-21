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

// Summary: NoOpToolManager is a no-op implementation of tool.ManagerInterface. A tool manager that does nothing.
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

// Summary: AddTool implements tool.ManagerInterface. No-op AddTool.
//
// Parameters:
//   - _ (tool.Tool): The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// Summary: GetTool implements tool.ManagerInterface. No-op GetTool.
//
// Parameters:
//   - _ (string): The _ parameter.
//
// Returns:
//   - tool.Tool: The resulting tool.Tool.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// Summary: ListTools implements tool.ManagerInterface. Returns an empty list of tools.
//
// Parameters:
//   - None.
//
// Returns:
//   - []tool.Tool: The resulting []tool.Tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// Summary: ListMCPTools implements tool.ManagerInterface. Returns an empty list of MCP tools.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*mcp.Tool: The resulting []*mcp.Tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// Summary: ClearToolsForService implements tool.ManagerInterface. No-op ClearToolsForService.
//
// Parameters:
//   - _ (string): The _ parameter.
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

// Summary: ExecuteTool implements tool.ManagerInterface. No-op ExecuteTool.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (*tool.ExecutionRequest): The _ parameter.
//
// Returns:
//   - any: The resulting any.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// Summary: SetMCPServer implements tool.ManagerInterface. No-op SetMCPServer.
//
// Parameters:
//   - _ (tool.MCPServerProvider): The _ parameter.
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

// Summary: AddMiddleware implements tool.ManagerInterface. No-op AddMiddleware.
//
// Parameters:
//   - _ (tool.ExecutionMiddleware): The _ parameter.
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

// Summary: AddServiceInfo implements tool.ManagerInterface. No-op AddServiceInfo.
//
// Parameters:
//   - _ (string): The _ parameter.
//   - _ (*tool.ServiceInfo): The _ parameter.
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

// Summary: GetServiceInfo implements tool.ManagerInterface. No-op GetServiceInfo.
//
// Parameters:
//   - _ (string): The _ parameter.
//
// Returns:
//   - *tool.ServiceInfo: The resulting *tool.ServiceInfo.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// Summary: ListServices implements tool.ManagerInterface. Returns an empty list of services.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*tool.ServiceInfo: The resulting []*tool.ServiceInfo.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// Summary: SetProfiles implements tool.ManagerInterface. No-op SetProfiles.
//
// Parameters:
//   - _ ([]string): The _ parameter.
//   - _ ([]*configv1.ProfileDefinition): The _ parameter.
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

// Summary: IsServiceAllowed implements tool.ManagerInterface. No-op IsServiceAllowed.
//
// Parameters:
//   - _ (string): The _ parameter.
//   - _ (string): The _ parameter.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// Summary: ToolMatchesProfile implements tool.ManagerInterface. No-op ToolMatchesProfile.
//
// Parameters:
//   - _ (tool.Tool): The _ parameter.
//   - _ (string): The _ parameter.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// Summary: GetAllowedServiceIDs implements tool.ManagerInterface. No-op GetAllowedServiceIDs.
//
// Parameters:
//   - _ (string): The _ parameter.
//
// Returns:
//   - map[string]bool: The resulting map[string]bool.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// Summary: GetToolCountForService implements tool.ManagerInterface. No-op GetToolCountForService.
//
// Parameters:
//   - _ (string): The _ parameter.
//
// Returns:
//   - int: The resulting int.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// Summary: NoOpPromptManager is a no-op implementation of prompt.ManagerInterface. A prompt manager that does nothing.
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

// Summary: AddPrompt implements prompt.ManagerInterface. No-op AddPrompt.
//
// Parameters:
//   - _ (prompt.Prompt): The _ parameter.
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

// Summary: UpdatePrompt implements prompt.ManagerInterface. No-op UpdatePrompt.
//
// Parameters:
//   - _ (prompt.Prompt): The _ parameter.
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

// Summary: GetPrompt implements prompt.ManagerInterface. No-op GetPrompt.
//
// Parameters:
//   - _ (string): The _ parameter.
//
// Returns:
//   - prompt.Prompt: The resulting prompt.Prompt.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// Summary: ListPrompts implements prompt.ManagerInterface. Returns an empty list of prompts.
//
// Parameters:
//   - None.
//
// Returns:
//   - []prompt.Prompt: The resulting []prompt.Prompt.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// Summary: ClearPromptsForService implements prompt.ManagerInterface. No-op ClearPromptsForService.
//
// Parameters:
//   - _ (string): The _ parameter.
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

// Summary: SetMCPServer implements prompt.ManagerInterface. No-op SetMCPServer.
//
// Parameters:
//   - _ (prompt.MCPServerProvider): The _ parameter.
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

// Summary: NoOpResourceManager is a no-op implementation of resource.ManagerInterface. A resource manager that does nothing.
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

// Summary: GetResource implements resource.ManagerInterface. No-op GetResource.
//
// Parameters:
//   - _ (string): The _ parameter.
//
// Returns:
//   - resource.Resource: The resulting resource.Resource.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// Summary: AddResource implements resource.ManagerInterface. No-op AddResource.
//
// Parameters:
//   - _ (resource.Resource): The _ parameter.
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

// Summary: RemoveResource implements resource.ManagerInterface. No-op RemoveResource.
//
// Parameters:
//   - _ (string): The _ parameter.
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

// Summary: ListResources implements resource.ManagerInterface. Returns an empty list of resources.
//
// Parameters:
//   - None.
//
// Returns:
//   - []resource.Resource: The resulting []resource.Resource.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// Summary: OnListChanged implements resource.ManagerInterface. No-op OnListChanged.
//
// Parameters:
//   - _ (func()): The _ parameter.
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

// Summary: ClearResourcesForService implements resource.ManagerInterface. No-op ClearResourcesForService.
//
// Parameters:
//   - _ (string): The _ parameter.
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
