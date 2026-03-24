// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// Summary: TemporaryToolManager is a tool manager that stores service info and tools temporarily. It is intended for use in ValidateService where we need to store service info and discovered tools for the duration of the validation request but discard them afterwards. Represents a TemporaryToolManager.
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

// Summary: NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TemporaryToolManager: The resulting *TemporaryToolManager.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
		tools:       make(map[string]tool.Tool),
	}
}

// Summary: AddServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//   - info (*tool.ServiceInfo): The info parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// Summary: GetServiceInfo implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - *tool.ServiceInfo: The resulting *tool.ServiceInfo.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// Summary: AddTool implements tool.ManagerInterface.
//
// Parameters:
//   - t (tool.Tool): The t parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetTool implements tool.ManagerInterface.
//
// Parameters:
//   - toolName (string): The toolName parameter.
//
// Returns:
//   - tool.Tool: The resulting tool.Tool.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *TemporaryToolManager) GetTool(toolName string) (tool.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil, false
	}
	t, ok := m.tools[toolName]
	return t, ok
}

// Summary: ListTools implements tool.ManagerInterface.
//
// Parameters:
//   - None.
//
// Returns:
//   - []tool.Tool: The resulting []tool.Tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: GetToolCountForService implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - int: The resulting int.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
