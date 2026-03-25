// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockToolManager is a mock of ToolManager.
type MockToolManager struct {
	mock.Mock
}

// CallTool is a mock method.
func (m *MockToolManager) CallTool(ctx context.Context, name string, args map[string]any) (any, error) {
	callArgs := m.Called(ctx, name, args)
	return callArgs.Get(0), callArgs.Error(1)
}

// ListTools is a mock method.
func (m *MockToolManager) ListTools(ctx context.Context) ([]any, error) {
	callArgs := m.Called(ctx)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]any), callArgs.Error(1)
}
