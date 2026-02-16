// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: Manages service info temporarily for validation.
//
// It is intended for use in ValidateService where we need to store service info
// for the duration of the validation request but discard it afterwards.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: Initializes a new TemporaryToolManager.
//
// Returns:
//   - *TemporaryToolManager: A new instance.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: Adds service information.
//
// Parameters:
//   - serviceID: string. The service ID.
//   - info: *tool.ServiceInfo. The service info.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: Retrieves service information.
//
// Parameters:
//   - serviceID: string. The service ID.
//
// Returns:
//   - *tool.ServiceInfo: The service info.
//   - bool: True if found.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: Returns the tool count (always 0).
//
// Parameters:
//   - _: string. Unused.
//
// Returns:
//   - int: 0.
func (m *TemporaryToolManager) GetToolCountForService(_ string) int {
	return 0
}
