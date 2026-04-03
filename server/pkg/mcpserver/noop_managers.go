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

// NoOpToolManager represents the public NoOpToolManager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the op tool manager components.
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

// AddTool serves as a public interface for interacting with AddTool.
//
// Summary: Add the tool appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) AddTool(_ tool.Tool) error { return nil }

// GetTool serves as a public interface for interacting with GetTool.
//
// Summary: Fetches and returns the underlying tool from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }

// ListTools serves as a public interface for interacting with ListTools.
//
// Summary: List the tools appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) ListTools() []tool.Tool { return nil }

// ListMCPTools serves as a public interface for interacting with ListMCPTools.
//
// Summary: List the mcp tools appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) ListMCPTools() []*mcp.Tool { return nil }

// ClearToolsForService serves as a public interface for interacting with ClearToolsForService.
//
// Summary: Clear the tools for service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) ClearToolsForService(_ string) {}

// ExecuteTool serves as a public interface for interacting with ExecuteTool.
//
// Summary: Execute the tool appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

// SetMCPServer serves as a public interface for interacting with SetMCPServer.
//
// Summary: Set the mcp server appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

// AddMiddleware serves as a public interface for interacting with AddMiddleware.
//
// Summary: Add the middleware appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}

// AddServiceInfo serves as a public interface for interacting with AddServiceInfo.
//
// Summary: Add the service info appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

// GetServiceInfo serves as a public interface for interacting with GetServiceInfo.
//
// Summary: Fetches and returns the underlying service info from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }

// ListServices serves as a public interface for interacting with ListServices.
//
// Summary: List the services appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) ListServices() []*tool.ServiceInfo { return nil }

// SetProfiles serves as a public interface for interacting with SetProfiles.
//
// Summary: Set the profiles appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

// IsServiceAllowed serves as a public interface for interacting with IsServiceAllowed.
//
// Summary: Checks condition indicating whether the target is service allowed.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) IsServiceAllowed(_, _ string) bool { return true }

// ToolMatchesProfile serves as a public interface for interacting with ToolMatchesProfile.
//
// Summary: Tool the matches profile appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) ToolMatchesProfile(_ tool.Tool, _ string) bool { return true }

// GetAllowedServiceIDs serves as a public interface for interacting with GetAllowedServiceIDs.
//
// Summary: Fetches and returns the underlying allowed service i ds from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, false
}

// GetToolCountForService serves as a public interface for interacting with GetToolCountForService.
//
// Summary: Fetches and returns the underlying tool count for service from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpToolManager) GetToolCountForService(_ string) int {
	return 0
}

// NoOpPromptManager represents the public NoOpPromptManager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the op prompt manager components.
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

// AddPrompt serves as a public interface for interacting with AddPrompt.
//
// Summary: Add the prompt appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpPromptManager) AddPrompt(_ prompt.Prompt) {}

// UpdatePrompt serves as a public interface for interacting with UpdatePrompt.
//
// Summary: Update the prompt appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpPromptManager) UpdatePrompt(_ prompt.Prompt) {}

// GetPrompt serves as a public interface for interacting with GetPrompt.
//
// Summary: Fetches and returns the underlying prompt from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) { return nil, false }

// ListPrompts serves as a public interface for interacting with ListPrompts.
//
// Summary: List the prompts appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpPromptManager) ListPrompts() []prompt.Prompt { return nil }

// ClearPromptsForService serves as a public interface for interacting with ClearPromptsForService.
//
// Summary: Clear the prompts for service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpPromptManager) ClearPromptsForService(_ string) {}

// SetMCPServer serves as a public interface for interacting with SetMCPServer.
//
// Summary: Set the mcp server appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

// NoOpResourceManager represents the public NoOpResourceManager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the op resource manager components.
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

// GetResource serves as a public interface for interacting with GetResource.
//
// Summary: Fetches and returns the underlying resource from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpResourceManager) GetResource(_ string) (resource.Resource, bool) { return nil, false }

// AddResource serves as a public interface for interacting with AddResource.
//
// Summary: Add the resource appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpResourceManager) AddResource(_ resource.Resource) {}

// RemoveResource serves as a public interface for interacting with RemoveResource.
//
// Summary: Remove the resource appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpResourceManager) RemoveResource(_ string) {}

// ListResources serves as a public interface for interacting with ListResources.
//
// Summary: List the resources appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpResourceManager) ListResources() []resource.Resource { return nil }

// OnListChanged serves as a public interface for interacting with OnListChanged.
//
// Summary: On the list changed appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpResourceManager) OnListChanged(_ func()) {}

// ClearResourcesForService serves as a public interface for interacting with ClearResourcesForService.
//
// Summary: Clear the resources for service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *NoOpResourceManager) ClearResourcesForService(_ string) {}
