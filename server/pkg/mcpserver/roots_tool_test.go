// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSession mocks the Session interface.
// MockSession mocks the Session interface.
// Summary: MockSession
	mock.Mock
}

// ListRoots ...
// Summary: ListRoots
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.ListRootsResult), args.Error(1)
}

// CreateMessage ...
// Summary: CreateMessage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.CreateMessageResult), args.Error(1)
}

// TestRootsTool_Execute ...
// Summary: TestRootsTool_Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Run("successful execution", func(t *testing.T) {
		rootsTool := NewRootsTool()
		ctx := context.Background()

		mockSession := new(MockSession)
		expectedRoots := &mcp.ListRootsResult{
			Roots: []*mcp.Root{
				{
					URI:  "file:///home/user",
					Name: "Home",
				},
			},
		}

		mockSession.On("ListRoots", mock.Anything).Return(expectedRoots, nil)

		// Create a context with the mock session
		ctx = tool.NewContextWithSession(ctx, mockSession)

		result, err := rootsTool.Execute(ctx, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedRoots, result)
		mockSession.AssertExpectations(t)
	})

	t.Run("no session in context", func(t *testing.T) {
		rootsTool := NewRootsTool()
		ctx := context.Background()

		result, err := rootsTool.Execute(ctx, nil)

		assert.Error(t, err)
		assert.Equal(t, "no active session available", err.Error())
		assert.Nil(t, result)
	})

	t.Run("failed to list roots", func(t *testing.T) {
		rootsTool := NewRootsTool()
		ctx := context.Background()

		mockSession := new(MockSession)
		mockSession.On("ListRoots", mock.Anything).Return((*mcp.ListRootsResult)(nil), errors.New("list error"))

		ctx = tool.NewContextWithSession(ctx, mockSession)

		result, err := rootsTool.Execute(ctx, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list roots")
		assert.Nil(t, result)
		mockSession.AssertExpectations(t)
	})
}

// TestRootsTool_Metadata ...
// Summary: TestRootsTool_Metadata
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	rootsTool := NewRootsTool()

	assert.NotNil(t, rootsTool.Tool())
	assert.NotNil(t, rootsTool.MCPTool())
	assert.Equal(t, "mcp:list_roots", rootsTool.Tool().GetName())
	assert.Nil(t, rootsTool.GetCacheConfig())
}

// TestRootsTool_Interface ...
// Summary: TestRootsTool_Interface
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	var _ tool.Tool = (*RootsTool)(nil)
}
