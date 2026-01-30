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
//   - _ (tool.Tool): The tool to add (unused).
//
// Returns:
//   - error: Always returns nil.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Parameters:
//   - _ (string): The name of the tool (unused).
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
//   - _ (string): The service ID (unused).
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Parameters:
//   - _ (context.Context): The context (unused).
//   - _ (*tool.ExecutionRequest): The execution request (unused).
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
//   - _ (tool.MCPServerProvider): The server provider (unused).
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Parameters:
//   - _ (tool.ExecutionMiddleware): The middleware to add (unused).
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - _ (string): The service ID (unused).
//   - _ (*tool.ServiceInfo): The service info (unused).
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - _ (string): The service ID (unused).
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
//   - _ ([]string): The profile names (unused).
//   - _ ([]*configv1.ProfileDefinition): The profile definitions (unused).
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Parameters:
//   - _ (string): The service ID (unused).
//   - _ (string): The profile ID (unused).
//
// Returns:
//   - bool: Always true.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Parameters:
//   - _ (tool.Tool): The tool (unused).
//   - _ (string): The profile ID (unused).
//
// Returns:
//   - bool: Always true.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Parameters:
//   - _ (string): The profile ID (unused).
//
// Returns:
//   - map[string]bool: Always nil.
//   - bool: Always false.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - _ (prompt.Prompt): The prompt to add (unused).
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - _ (prompt.Prompt): The prompt to update (unused).
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - _ (string): The name of the prompt (unused).
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
//   - _ (string): The service ID (unused).
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Parameters:
//   - _ (prompt.MCPServerProvider): The server provider (unused).
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Parameters:
//   - _ (string): The resource URI (unused).
//
// Returns:
//   - resource.Resource: Always nil.
//   - bool: Always false.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Parameters:
//   - _ (resource.Resource): The resource to add (unused).
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Parameters:
//   - _ (string): The resource URI (unused).
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Returns:
//   - []resource.Resource: Always nil.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Parameters:
//   - _ (func()): The callback function (unused).
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Parameters:
//   - _ (string): The service ID (unused).
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
