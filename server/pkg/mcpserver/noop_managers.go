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

// Summary: NoOpToolManager represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Summary: AddTool executes the operation.
//
// Parameters:
//   - _ tool.Tool: Input parameter.
//
// Returns:
//   - error { return nil }
: Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Summary: No-op GetTool.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
// Summary: GetTool executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - (tool.Tool, bool): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Summary: Returns an empty list of tools.
//
// Parameters:
//   - None.
//
// Returns:
//   - []tool.Tool: Always nil.
// Summary: ListTools executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - []tool.Tool { return nil }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Summary: Returns an empty list of MCP tools.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*mcp.Tool: Always nil.
//
// Side Effects:
// Summary: ListMCPTools executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - []*mcp.Tool { return nil }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Summary: No-op ClearToolsForService.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: ClearToolsForService executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Summary: No-op ExecuteTool.
//
// Parameters:
//   - _ (context.Context): Unused.
//   - _ (*tool.ExecutionRequest): Unused.
//
// Returns:
//   - any: Always nil.
//   - error: Always nil.
// Summary: ExecuteTool executes the operation.
//
// Parameters:
//   - _ context.Context: Input parameter.
//   - _ *tool.ExecutionRequest: Input parameter.
//
// Returns:
//   - (any, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Summary: No-op SetMCPServer.
//
// Parameters:
//   - _ (tool.MCPServerProvider): Unused.
//
// Returns:
//   - None.
//
// Summary: SetMCPServer executes the operation.
//
// Parameters:
//   - _ tool.MCPServerProvider: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Summary: No-op AddMiddleware.
//
// Parameters:
//   - _ (tool.ExecutionMiddleware): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: AddMiddleware executes the operation.
//
// Parameters:
//   - _ tool.ExecutionMiddleware: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: No-op AddServiceInfo.
//
// Parameters:
//   - _ (string): Unused.
//   - _ (*tool.ServiceInfo): Unused.
//
// Returns:
//   - None.
//
// Summary: AddServiceInfo executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//   - _ *tool.ServiceInfo: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: No-op GetServiceInfo.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - *tool.ServiceInfo: Always nil.
//   - bool: Always false.
// Summary: GetServiceInfo executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - (*tool.ServiceInfo, bool): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Summary: Returns an empty list of services.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*tool.ServiceInfo: Always nil.
//
// Side Effects:
// Summary: ListServices executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - []*tool.ServiceInfo { return nil }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Summary: No-op SetProfiles.
//
// Parameters:
//   - _ ([]string): Unused.
//   - _ ([]*configv1.ProfileDefinition): Unused.
//
// Returns:
//   - None.
//
// Summary: SetProfiles executes the operation.
//
// Parameters:
//   - _ []string: Input parameter.
//   - _ []*configv1.ProfileDefinition: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Summary: No-op IsServiceAllowed.
//
// Parameters:
//   - _, _ (string): Unused.
//
// Returns:
//   - bool: Always true (allow all).
// Summary: IsServiceAllowed executes the operation.
//
// Parameters:
//   - _: Input parameter.
//   - _ string: Input parameter.
//
// Returns:
//   - bool { return true }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Summary: No-op ToolMatchesProfile.
//
// Parameters:
//   - _ (tool.Tool): Unused.
//   - _ (string): Unused.
//
// Returns:
//   - bool: Always true.
// Summary: ToolMatchesProfile executes the operation.
//
// Parameters:
//   - _ tool.Tool: Input parameter.
//   - _ string: Input parameter.
//
// Returns:
//   - bool { return true }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Summary: No-op GetAllowedServiceIDs.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - map[string]bool: Always nil.
//   - bool: Always false.
// Summary: GetAllowedServiceIDs executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - (map[string]bool, bool): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: No-op GetToolCountForService.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - int: Always 0.
// Summary: GetToolCountForService executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - int {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// Summary: A prompt manager that does nothing.
// Summary: NoOpPromptManager represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Summary: No-op AddPrompt.
//
// Parameters:
//   - _ (prompt.Prompt): Unused.
//
// Summary: AddPrompt executes the operation.
//
// Parameters:
//   - _ prompt.Prompt: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Summary: No-op UpdatePrompt.
//
// Parameters:
//   - _ (prompt.Prompt): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: UpdatePrompt executes the operation.
//
// Parameters:
//   - _ prompt.Prompt: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Summary: No-op GetPrompt.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - prompt.Prompt: Always nil.
//   - bool: Always false.
// Summary: GetPrompt executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - (prompt.Prompt, bool): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Summary: Returns an empty list of prompts.
//
// Parameters:
//   - None.
//
// Returns:
//   - []prompt.Prompt: Always nil.
//
// Summary: ListPrompts executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - []prompt.Prompt { return nil }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Summary: No-op ClearPromptsForService.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: ClearPromptsForService executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Summary: No-op SetMCPServer.
//
// Parameters:
//   - _ (prompt.MCPServerProvider): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: SetMCPServer executes the operation.
//
// Parameters:
//   - _ prompt.MCPServerProvider: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary: A resource manager that does nothing.
// Summary: NoOpResourceManager represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Summary: No-op GetResource.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - resource.Resource: Always nil.
//   - bool: Always false.
// Summary: GetResource executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - (resource.Resource, bool): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Summary: No-op AddResource.
//
// Parameters:
//   - _ (resource.Resource): Unused.
//
// Returns:
//   - None.
// Summary: AddResource executes the operation.
//
// Parameters:
//   - _ resource.Resource: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Summary: No-op RemoveResource.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: RemoveResource executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Summary: Returns an empty list of resources.
//
// Parameters:
//   - None.
//
// Returns:
//   - []resource.Resource: Always nil.
//
// Side Effects:
// Summary: ListResources executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - []resource.Resource { return nil }
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Summary: No-op OnListChanged.
//
// Parameters:
//   - _ (func()): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: OnListChanged executes the operation.
//
// Parameters:
//   - _ func(: Input parameter.
//
// Returns:
//   - ): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Summary: No-op ClearResourcesForService.
//
// Parameters:
//   - _ (string): Unused.
//
// Returns:
//   - None.
//
// Side Effects:
// Summary: ClearResourcesForService executes the operation.
//
// Parameters:
//   - _ string: Input parameter.
//
// Returns:
//   - {}
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
