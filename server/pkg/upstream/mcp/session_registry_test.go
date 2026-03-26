// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"testing"

	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestSessionRegistry(t *testing.T) {
	registry := NewSessionRegistry()
	mockDownstream := new(MockSession)
	mockUpstream := &mcp_sdk.ServerSession{} // Using ServerSession as a dummy implementer of mcp.Session (interface check)

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

func TestSessionRegistry_Concurrency(t *testing.T) {
	registry := NewSessionRegistry()
	mockDownstream := new(MockSession)
	mockUpstream := &mcp_sdk.ServerSession{}

	const routines = 100
	done := make(chan struct{})

	for i := 0; i < routines; i++ {
		go func() {
			registry.Register(mockUpstream, mockDownstream)
			registry.Get(mockUpstream)
			registry.Unregister(mockUpstream)
			done <- struct{}{}
		}()
	}

	for i := 0; i < routines; i++ {
		<-done
	}

	// Should be empty or not, but at least shouldn't crash
	_, _ = registry.Get(mockUpstream)
}

func TestSessionRegistry_MultipleSessions(t *testing.T) {
	registry := NewSessionRegistry()

	mockUpstream1 := &mcp_sdk.ServerSession{}
	mockDownstream1 := new(MockSession)

	mockUpstream2 := &mcp_sdk.ServerSession{} // Go uses pointers, so another &mcp.ServerSession{} has a different memory address
	mockDownstream2 := new(MockSession)

	registry.Register(mockUpstream1, mockDownstream1)
	registry.Register(mockUpstream2, mockDownstream2)

	got1, ok1 := registry.Get(mockUpstream1)
	require.True(t, ok1)
	require.Equal(t, mockDownstream1, got1)

	got2, ok2 := registry.Get(mockUpstream2)
	require.True(t, ok2)
	require.Equal(t, mockDownstream2, got2)

	registry.Unregister(mockUpstream1)

	_, ok1 = registry.Get(mockUpstream1)
	require.False(t, ok1)

	got2, ok2 = registry.Get(mockUpstream2)
	require.True(t, ok2)
	require.Equal(t, mockDownstream2, got2)
}
