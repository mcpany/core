// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// TemporaryToolManager represents the public TemporaryToolManager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the tool manager components.
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
type TemporaryToolManager struct {
	NoOpToolManager
	mu          sync.RWMutex
	serviceInfo map[string]*tool.ServiceInfo
	tools       map[string]tool.Tool
}

// NewTemporaryToolManager serves as a public interface for interacting with NewTemporaryToolManager.
//
// Summary: Constructs and returns an initialized temporary tool manager ready for consumption.
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
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
		tools:       make(map[string]tool.Tool),
	}
}

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
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

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
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

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
func (m *TemporaryToolManager) AddTool(t tool.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tools == nil {
		m.tools = make(map[string]tool.Tool)
	}

	if t.Tool().GetServiceId() == "" {
		return fmt.Errorf("tool service ID cannot be empty")
	}

	sanitizedToolName, err := util.SanitizeToolName(t.Tool().GetName())
	if err != nil {
		return fmt.Errorf("failed to sanitize tool name: %w", err)
	}

	toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
	m.tools[toolID] = t
	return nil
}

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
func (m *TemporaryToolManager) GetTool(toolName string) (tool.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil, false
	}
	t, ok := m.tools[toolName]
	return t, ok
}

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
func (m *TemporaryToolManager) ListTools() []tool.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil
	}
	list := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		list = append(list, t)
	}
	return list
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
func (m *TemporaryToolManager) GetToolCountForService(serviceID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return 0
	}
	count := 0
	for _, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			count++
		}
	}
	return count
}
