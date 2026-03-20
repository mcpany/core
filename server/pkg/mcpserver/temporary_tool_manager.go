// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// TemporaryToolManager is a tool manager that stores service info and tools temporarily.
//
// Summary: TemporaryToolManager is a tool manager that stores service info and tools temporarily.
type TemporaryToolManager struct {
	NoOpToolManager
	mu          sync.RWMutex
	serviceInfo map[string]*tool.ServiceInfo
	tools       map[string]tool.Tool
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TemporaryToolManager: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
		tools:       make(map[string]tool.Tool),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: AddServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID (string): The textual representation of serviceid.
//   - info (*tool.ServiceInfo): The provided info data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: GetServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID (string): The textual representation of serviceid.
//
// Returns:
//   - *tool.ServiceInfo: The resulting object or data structure.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// AddTool implements tool.ManagerInterface.
//
// Summary: AddTool implements tool.ManagerInterface.
//
// Parameters:
//   - t (tool.Tool): The provided t data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// GetTool implements tool.ManagerInterface.
//
// Summary: GetTool implements tool.ManagerInterface.
//
// Parameters:
//   - toolName (string): The human-readable or system name.
//
// Returns:
//   - tool.Tool: The resulting object or data structure.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *TemporaryToolManager) GetTool(toolName string) (tool.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil, false
	}
	t, ok := m.tools[toolName]
	return t, ok
}

// ListTools implements tool.ManagerInterface.
//
// Summary: ListTools implements tool.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []tool.Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: GetToolCountForService implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID (string): The textual representation of serviceid.
//
// Returns:
//   - int: The calculated numeric value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
