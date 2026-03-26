// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSession mocks the tool.Session interface.
// MockSession mocks the tool.Session interface.
// Summary: MockSession
	mock.Mock
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
	return args.Get(0).(*mcp.CreateMessageResult), args.Error(1)
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
	return args.Get(0).(*mcp.ListRootsResult), args.Error(1)
}

// TestSessionRegistry ...
// Summary: TestSessionRegistry
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	registry := NewSessionRegistry()
	mockDownstream := new(MockSession)
	mockUpstream := &mcp.ServerSession{} // Using ServerSession as a dummy implementer of mcp.Session (interface check)

	// Register
	registry.Register(mockUpstream, mockDownstream)

	// Get
	got, ok := registry.Get(mockUpstream)
	require.True(t, ok)
	require.Equal(t, mockDownstream, got)

	// Unregister
	registry.Unregister(mockUpstream)

	// Get after unregister
	_, ok = registry.Get(mockUpstream)
	require.False(t, ok)
}
