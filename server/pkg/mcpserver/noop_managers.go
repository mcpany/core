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
// Summary. A tool manager that does nothing.
type NoOpToolManager struct{}

// AddTool provides addtool functionality.
//
// Summary: AddTool.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool provides gettool functionality.
//
// Summary: GetTool.
//
// Parameters.
//   - _: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools provides listtools functionality.
//
// Summary: ListTools.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools provides listmcptools functionality.
//
// Summary: ListMCPTools.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService provides cleartoolsforservice functionality.
//
// Summary: ClearToolsForService.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool provides executetool functionality.
//
// Summary: ExecuteTool.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer provides setmcpserver functionality.
//
// Summary: SetMCPServer.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware provides addmiddleware functionality.
//
// Summary: AddMiddleware.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo provides addserviceinfo functionality.
//
// Summary: AddServiceInfo.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo provides getserviceinfo functionality.
//
// Summary: GetServiceInfo.
//
// Parameters.
//   - _: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices provides listservices functionality.
//
// Summary: ListServices.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles provides setprofiles functionality.
//
// Summary: SetProfiles.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed provides isserviceallowed functionality.
//
// Summary: IsServiceAllowed.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile provides toolmatchesprofile functionality.
//
// Summary: ToolMatchesProfile.
//
// Parameters.
//   - _: The parameter.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs provides getallowedserviceids functionality.
//
// Summary: GetAllowedServiceIDs.
//
// Parameters.
//   - _: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService provides gettoolcountforservice functionality.
//
// Summary: GetToolCountForService.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager is a no-op implementation of prompt.ManagerInterface.
//
// Summary. A prompt manager that does nothing.
type NoOpPromptManager struct{}

// AddPrompt provides addprompt functionality.
//
// Summary: AddPrompt.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt provides updateprompt functionality.
//
// Summary: UpdatePrompt.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt provides getprompt functionality.
//
// Summary: GetPrompt.
//
// Parameters.
//   - _: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts provides listprompts functionality.
//
// Summary: ListPrompts.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService provides clearpromptsforservice functionality.
//
// Summary: ClearPromptsForService.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer provides setmcpserver functionality.
//
// Summary: SetMCPServer.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager is a no-op implementation of resource.ManagerInterface.
//
// Summary. A resource manager that does nothing.
type NoOpResourceManager struct{}

// GetResource provides getresource functionality.
//
// Summary: GetResource.
//
// Parameters.
//   - _: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource provides addresource functionality.
//
// Summary: AddResource.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource provides removeresource functionality.
//
// Summary: RemoveResource.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources provides listresources functionality.
//
// Summary: ListResources.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged provides onlistchanged functionality.
//
// Summary: OnListChanged.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService provides clearresourcesforservice functionality.
//
// Summary: ClearResourcesForService.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - None.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
