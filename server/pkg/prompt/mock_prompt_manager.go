// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockPromptManager is a mock of PromptManager.
type MockPromptManager struct {
	mock.Mock
}

// GetPrompt is a mock method.
func (m *MockPromptManager) GetPrompt(ctx context.Context, id string) (any, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

// ListPrompts is a mock method.
func (m *MockPromptManager) ListPrompts(ctx context.Context) ([]any, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}
