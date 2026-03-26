// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0.

package resource

import (
	"github.com/stretchr/testify/mock"
)

// MockResourceManager is a mock of ManagerInterface.
type MockResourceManager struct {
	mock.Mock
}

// GetResource is a mock method.
func (m *MockResourceManager) GetResource(uri string) (Resource, bool) {
	args := m.Called(uri)
	return args.Get(0).(Resource), args.Bool(1)
}

// AddResource is a mock method.
func (m *MockResourceManager) AddResource(res Resource) {
	m.Called(res)
}

// RemoveResource is a mock method.
func (m *MockResourceManager) RemoveResource(uri string) {
	m.Called(uri)
}

// ListResources is a mock method.
func (m *MockResourceManager) ListResources() []Resource {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]Resource)
}

// OnListChanged is a mock method.
func (m *MockResourceManager) OnListChanged(f func()) {
	m.Called(f)
}

// ClearResourcesForService is a mock method.
func (m *MockResourceManager) ClearResourcesForService(serviceID string) {
	m.Called(serviceID)
}
