// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockResourceManager is a mock of ResourceManager.
type MockResourceManager struct {
	mock.Mock
}

// GetResource is a mock method.
func (m *MockResourceManager) GetResource(ctx context.Context, id string) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

// ListResources is a mock method.
func (m *MockResourceManager) ListResources(ctx context.Context) ([]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).([]interface{}), args.Error(1)
}
