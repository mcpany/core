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
//
// Parameters:
//   - _: tool.Tool. The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   - Returns an error if the operation fails.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
//
// Parameters:
//   - _: string. The _ parameter.
//
// Returns:
//   - tool.Tool: The result.
//   - bool: The result.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Returns the result.
//
// Returns:
//   - []tool.Tool: The result.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Returns the result.
//
// Returns:
//   - []*mcp.Tool: The result.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: string. The _ parameter.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//   - _: context.Context. The context for the operation.
//   - _: The request object.
//
// Returns:
//   - any: The result.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   - Returns an error if the operation fails.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: tool.MCPServerProvider. The _ parameter.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: tool.ExecutionMiddleware. The _ parameter.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Parameters:
//   - _: string. The _ parameter.
//   - _: *tool.ServiceInfo. The tool.ServiceInfo instance.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
//
// Parameters:
//   - _: string. The _ parameter.
//
// Returns:
//   - *tool.ServiceInfo: The result.
//   - bool: The result.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Returns the result.
//
// Returns:
//   - []*tool.ServiceInfo: The result.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Parameters:
//   - _: []string. A list of strings.
//   - _: []*configv1.ProfileDefinition. A list of *configv1.ProfileDefinitions.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns true if successful.
//
// Parameters:
//   - _: string. The _ parameter.
//
// Returns:
//   - bool: The result.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns true if successful.
//
// Parameters:
//   - _: tool.Tool. The _ parameter.
//   - _: string. The _ parameter.
//
// Returns:
//   - bool: The result.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
//
// Parameters:
//   - _: string. The _ parameter.
//
// Returns:
//   - map[string]bool: The result.
//   - bool: The result.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: prompt.Prompt. The _ parameter.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: prompt.Prompt. The _ parameter.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
//
// Parameters:
//   - _: string. The _ parameter.
//
// Returns:
//   - prompt.Prompt: The result.
//   - bool: The result.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Returns the result.
//
// Returns:
//   - []prompt.Prompt: The result.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: string. The _ parameter.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: prompt.MCPServerProvider. The _ parameter.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns true if successful.
//
// Parameters:
//   - _: string. The _ parameter.
//
// Returns:
//   - resource.Resource: The result.
//   - bool: The result.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: resource.Resource. The _ parameter.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: string. The _ parameter.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Returns the result.
//
// Returns:
//   - []resource.Resource: The result.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: func(. The _ parameter.
//
// Returns:
//   - ): The result.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// _ is an unused parameter.
//
// Parameters:
//   - _: string. The _ parameter.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
