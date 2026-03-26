// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
<<<<<<< HEAD
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSession mocks the tool.Session interface
type MockSession struct {
	mock.Mock
}

func (m *MockSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	args := m.Called(ctx, params)
	if res := args.Get(0); res != nil {
		return res.(*mcp.CreateMessageResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSession) SessionID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSession) GetUpstreamSession() mcp.Session {
	args := m.Called()
	if res := args.Get(0); res != nil {
		return res.(mcp.Session)
	}
	return nil
}

func (m *MockSession) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	args := m.Called(ctx)
	if res := args.Get(0); res != nil {
		return res.(*mcp.ListRootsResult), args.Error(1)
	}
	return nil, args.Error(1)
}

=======
	"github.com/stretchr/testify/require"
)

>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func TestHandleCreateMessage(t *testing.T) {
	u := NewUpstream(configv1.GlobalSettings_builder{}.Build()).(*Upstream)
	ctx := context.Background()

	// Mock downstream session
	mockDownstream := new(MockSession)

	// Mock upstream session
	// We need a distinct instance for identifying the session
	mockUpstream := &mcp.ClientSession{}

	t.Run("Success", func(t *testing.T) {
		// Register session
		u.sessionRegistry.Register(mockUpstream, mockDownstream)
		defer u.sessionRegistry.Unregister(mockUpstream)

		params := &mcp.CreateMessageParams{
			Messages: []*mcp.SamplingMessage{},
		}

		expectedResult := &mcp.CreateMessageResult{
			Content: &mcp.TextContent{Text: "result"},
			Model:   "test-model",
			Role:    "assistant",
		}

		// Setup expectation
		mockDownstream.On("CreateMessage", ctx, params).Return(expectedResult, nil).Once()

		req := &mcp.ClientRequest[*mcp.CreateMessageParams]{
			Params:  params,
			Session: mockUpstream,
		}

		result, err := u.handleCreateMessage(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockDownstream.AssertExpectations(t)
	})

	t.Run("NoSessionInRequest", func(t *testing.T) {
		req := &mcp.ClientRequest[*mcp.CreateMessageParams]{
			Params:  &mcp.CreateMessageParams{Messages: []*mcp.SamplingMessage{}},
			Session: nil,
		}

		result, err := u.handleCreateMessage(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no session associated")
	})

	t.Run("NoDownstreamSessionFound", func(t *testing.T) {
		otherUpstream := &mcp.ClientSession{} // Different pointer

		req := &mcp.ClientRequest[*mcp.CreateMessageParams]{
			Params:  &mcp.CreateMessageParams{Messages: []*mcp.SamplingMessage{}},
			Session: otherUpstream,
		}

		result, err := u.handleCreateMessage(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no downstream session found")
	})
}
