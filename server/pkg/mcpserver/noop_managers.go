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
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Returns the result.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Returns the result.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Returns the result.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns true if successful.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns true if successful.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the count.
func (m *NoOpToolManager) GetToolCountForService(_ string) int { return 0 }

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Returns the result.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Returns the result.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// _ is an unused parameter.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
