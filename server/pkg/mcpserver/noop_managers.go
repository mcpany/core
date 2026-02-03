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
// It is useful for testing or for situations where tool management is not required.
type NoOpToolManager struct{}

// AddTool implements tool.ManagerInterface.
//
// Parameters:
//   - t: tool.Tool. The tool to add (ignored).
//
// Returns:
//   - error: Always returns nil.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool implements tool.ManagerInterface.
//
// Parameters:
//   - toolName: string. The name of the tool (ignored).
//
// Returns:
//   - tool.Tool: Always returns nil.
//   - bool: Always returns false.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools implements tool.ManagerInterface.
//
// Returns:
//   - []tool.Tool: Always returns nil.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools implements tool.ManagerInterface.
//
// Returns:
//   - []*mcp.Tool: Always returns nil.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID (ignored).
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool implements tool.ManagerInterface.
//
// Parameters:
//   - ctx: context.Context. The context (ignored).
//   - req: *tool.ExecutionRequest. The request (ignored).
//
// Returns:
//   - any: Always returns nil.
//   - error: Always returns nil.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer implements tool.ManagerInterface.
//
// Parameters:
//   - mcpServer: tool.MCPServerProvider. The provider (ignored).
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware implements tool.ManagerInterface.
//
// Parameters:
//   - middleware: tool.ExecutionMiddleware. The middleware (ignored).
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID (ignored).
//   - info: *tool.ServiceInfo. The info (ignored).
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID (ignored).
//
// Returns:
//   - *tool.ServiceInfo: Always returns nil.
//   - bool: Always returns false.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices implements tool.ManagerInterface.
//
// Returns:
//   - []*tool.ServiceInfo: Always returns nil.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles implements tool.ManagerInterface.
//
// Parameters:
//   - profiles: []string. The profiles (ignored).
//   - defs: []*configv1.ProfileDefinition. The definitions (ignored).
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID (ignored).
//   - profileID: string. The profile ID (ignored).
//
// Returns:
//   - bool: Always returns true.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile implements tool.ManagerInterface.
//
// Parameters:
//   - t: tool.Tool. The tool (ignored).
//   - profileID: string. The profile ID (ignored).
//
// Returns:
//   - bool: Always returns true.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs implements tool.ManagerInterface.
//
// Parameters:
//   - profileID: string. The profile ID (ignored).
//
// Returns:
//   - map[string]bool: Always returns nil.
//   - bool: Always returns false.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// It is useful for testing or for situations where prompt management is not required.
type NoOpPromptManager struct{}

// AddPrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - p: prompt.Prompt. The prompt (ignored).
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - p: prompt.Prompt. The prompt (ignored).
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt implements prompt.ManagerInterface.
//
// Parameters:
//   - name: string. The name (ignored).
//
// Returns:
//   - prompt.Prompt: Always returns nil.
//   - bool: Always returns false.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts implements prompt.ManagerInterface.
//
// Returns:
//   - []prompt.Prompt: Always returns nil.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService implements prompt.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID (ignored).
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer implements prompt.ManagerInterface.
//
// Parameters:
//   - s: prompt.MCPServerProvider. The provider (ignored).
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// It is useful for testing or for situations where resource management is not required.
type NoOpResourceManager struct{}

// GetResource implements resource.ManagerInterface.
//
// Parameters:
//   - uri: string. The URI (ignored).
//
// Returns:
//   - resource.Resource: Always returns nil.
//   - bool: Always returns false.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource implements resource.ManagerInterface.
//
// Parameters:
//   - r: resource.Resource. The resource (ignored).
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource implements resource.ManagerInterface.
//
// Parameters:
//   - uri: string. The URI (ignored).
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources implements resource.ManagerInterface.
//
// Returns:
//   - []resource.Resource: Always returns nil.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged implements resource.ManagerInterface.
//
// Parameters:
//   - f: func(). The callback (ignored).
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService implements resource.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The service ID (ignored).
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
