// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: Stores service information temporarily for validation purposes.
//
// It is intended for use in ValidateService where we need to store service info
// for the duration of the validation request but discard it afterwards.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: Initializes a new TemporaryToolManager instance.
//
// Returns:
//   - *TemporaryToolManager: A pointer to the new TemporaryToolManager.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo adds service information to the manager.
//
// Summary: Stores service information for a given service ID.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - info: *tool.ServiceInfo. The service information to store.
//
// Side Effects:
//   - Updates the internal service info map.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves service information by ID.
//
// Summary: Retrieves the stored service information for a given service ID.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *tool.ServiceInfo: The service information if found.
//   - bool: True if the service info exists, false otherwise.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// GetToolCountForService returns the number of tools for a service.
//
// Summary: Returns the tool count for a service (always 0 for this implementation).
//
// Parameters:
//   - _ : string. The service ID (unused).
//
// Returns:
//   - int: Always returns 0.
func (m *TemporaryToolManager) GetToolCountForService(_ string) int {
	return 0
}
