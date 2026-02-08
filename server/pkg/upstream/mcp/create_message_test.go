package mcp

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
