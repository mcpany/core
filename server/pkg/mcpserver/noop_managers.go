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
type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Parameters:
//   - t: tool.Tool. The tool to add.
//
// Returns:
//   - error: Always returns nil.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Parameters:
//   - name: string. The name of the tool.
//
// Returns:
//   - tool.Tool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Returns:
//   - []tool.Tool: Always nil.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Returns:
//   - []*mcp.Tool: Always nil.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Parameters:
//   - ctx: context.Context. The context.
//   - req: *tool.ExecutionRequest. The execution request.
//
// Returns:
//   - any: Always nil.
//   - error: Always nil.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Parameters:
//   - p: tool.MCPServerProvider. The provider.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Parameters:
//   - mw: tool.ExecutionMiddleware. The middleware.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
//   - info: *tool.ServiceInfo. The service info.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
//
// Returns:
//   - *tool.ServiceInfo: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Returns:
//   - []*tool.ServiceInfo: Always nil.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Parameters:
//   - profiles: []string. The profiles.
//   - definitions: []*configv1.ProfileDefinition. The definitions.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
//   - profileID: string. The profile ID.
//
// Returns:
//   - bool: Always true.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Parameters:
//   - t: tool.Tool. The tool.
//   - profileID: string. The profile ID.
//
// Returns:
//   - bool: Always true.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Parameters:
//   - profileID: string. The profile ID.
//
// Returns:
//   - map[string]bool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
//
// Returns:
//   - int: Always 0.
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - p: prompt.Prompt. The prompt to add.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - p: prompt.Prompt. The prompt to update.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - name: string. The name of the prompt.
//
// Returns:
//   - prompt.Prompt: Always nil.
//   - bool: Always false.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Returns:
//   - []prompt.Prompt: Always nil.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Parameters:
//   - p: prompt.MCPServerProvider. The provider.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Parameters:
//   - uri: string. The resource URI.
//
// Returns:
//   - resource.Resource: Always nil.
//   - bool: Always false.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Parameters:
//   - r: resource.Resource. The resource to add.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Parameters:
//   - uri: string. The resource URI.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Returns:
//   - []resource.Resource: Always nil.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Parameters:
//   - f: func(). The callback.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
