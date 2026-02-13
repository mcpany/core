// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: A temporary tool manager for validation requests.
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
//   - *TemporaryToolManager: The initialized manager.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: Stores service info temporarily.
//
// Parameters:
//   - serviceID: string. The service ID.
//   - info: *tool.ServiceInfo. The service info to store.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: Retrieves temporarily stored service info.
//
// Parameters:
//   - serviceID: string. The service ID.
//
// Returns:
//   - *tool.ServiceInfo: The service info if found.
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
// Summary: Returns tool count (always 0 for temporary manager).
//
// Parameters:
//   - _ : string. Unused.
//
// Returns:
//   - int: Always 0.
func (m *TemporaryToolManager) GetToolCountForService(_ string) int {
	return 0
}
