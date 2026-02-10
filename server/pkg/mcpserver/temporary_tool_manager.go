// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: A tool manager that stores service info temporarily.
//
// It is intended for use in ValidateService where we need to store service info
// for the duration of the validation request but discard it afterwards.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: Creates a new TemporaryToolManager.
//
// Returns:
//   - *TemporaryToolManager: A new instance of TemporaryToolManager.
//
// Throws/Errors:
//   - None.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: Adds service info to the temporary manager.
//
// Parameters:
//   - serviceID: string. The ID of the service.
//   - info: *tool.ServiceInfo. The service information.
//
// Returns:
//   None.
//
// Throws/Errors:
//   - None.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: Retrieves service info from the temporary manager.
//
// Parameters:
//   - serviceID: string. The ID of the service.
//
// Returns:
//   - *tool.ServiceInfo: The service information if found.
//   - bool: True if found, false otherwise.
//
// Throws/Errors:
//   - None.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: Returns the tool count for a service (always 0).
//
// Parameters:
//   - _: string. Service ID (unused).
//
// Returns:
//   - int: Always 0.
//
// Throws/Errors:
//   - None.
func (m *TemporaryToolManager) GetToolCountForService(_ string) int {
	return 0
}
