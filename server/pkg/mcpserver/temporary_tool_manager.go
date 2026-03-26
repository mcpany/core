// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// TemporaryToolManager temporaryToolManager represents a temporary tool manager.
//
// Summary: TemporaryToolManager represents a temporary tool manager.
type TemporaryToolManager struct {
	NoOpToolManager
	mu          sync.RWMutex
	serviceInfo map[string]*tool.ServiceInfo
	tools       map[string]tool.Tool
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Returns: - None.
//   - *TemporaryToolManager: A new instance of TemporaryToolManager.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewTemporaryToolManager operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
		tools:       make(map[string]tool.Tool),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Parameters: - None.
//   - serviceID (string): The ID of the service.
//   - info (*tool.ServiceInfo): The service information.
//
// Side Effects: - None.
//   - Updates the internal service info map.
//
// Summary: Executes AddServiceInfo operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
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
// Parameters: - None.
//   - serviceID (string): The ID of the service.
//
// Returns: - None.
//   - *tool.ServiceInfo: The service information if found.
//   - bool: True if the service information exists.
//
// Side Effects: - None.
//   - None.
//
// Summary: Retrieves GetServiceInfo operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// AddTool implements tool.ManagerInterface.
//
// Parameters: - None.
//   - t (tool.Tool): The tool to add.
//
// Returns: - None.
//   - error: An error if the tool service ID is empty or name sanitization fails.
//
// Side Effects: - None.
//   - Updates the internal tool map.
//
// Summary: Executes AddTool operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// GetTool implements tool.ManagerInterface.
//
// Parameters: - None.
//   - toolName (string): The name of the tool.
//
// Returns: - None.
//   - tool.Tool: The tool if found.
//   - bool: True if the tool exists.
//
// Side Effects: - None.
//   - None.
//
// Summary: Retrieves GetTool operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// ListTools implements tool.ManagerInterface.
//
// Returns: - None.
//   - []tool.Tool: A list of all tools.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes ListTools operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// GetToolCountForService implements tool.ManagerInterface.
//
// Parameters: - None.
//   - serviceID (string): The ID of the service.
//
// Returns: - None.
//   - int: The number of tools for the service.
//
// Side Effects: - None.
//   - None.
//
// Summary: Retrieves GetToolCountForService operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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
