package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSession mocks the tool.Session interface.
type MockSession struct {
	mock.Mock
}

func (m *MockSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*mcp.CreateMessageResult), args.Error(1)
}

func (m *MockSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	args := m.Called(ctx)
	return args.Get(0).(*mcp.ListRootsResult), args.Error(1)
}

func TestSessionRegistry(t *testing.T) {
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
