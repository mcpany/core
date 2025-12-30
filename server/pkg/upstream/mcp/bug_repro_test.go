// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestHandleCreateMessage_NilSession(t *testing.T) {
	// This test reproduces the bug where `session == nil` check fails for a nil *ClientSession wrapped in an interface.

	u := &Upstream{
		sessionRegistry: NewSessionRegistry(),
	}

	// We manually construct a ClientRequest with a nil Session.
	// Since req.Session is *ClientSession, and we initialize it to nil.
	req := &mcp.ClientRequest[*mcp.CreateMessageParams]{
		Session: nil,
		Params:  &mcp.CreateMessageParams{},
	}

	// req.GetSession() will return (Session)(nil), which is a nil interface if Session is just interface.
	// But in go-sdk, GetSession returns `r.Session`.
	// If r.Session is nil, it returns `(*ClientSession)(nil)` wrapped in interface.
	// This interface value is NOT nil.

	// When we call handleCreateMessage, it checks `if session == nil`.
	// This check will be false because the interface is not nil.
	// Then it calls `u.sessionRegistry.Get(session)`.
	// `sessionRegistry` map key is `mcp.Session`.
	// If the session is `(*ClientSession)(nil)`, it might be found or not found depending on registration.
	// But it shouldn't proceed to registry lookup if session is effectively nil.

	// Expectation: The function should return error "no session associated with request".
	// Actual (Bug): It proceeds and returns "no downstream session found for upstream session" (because registry lookup fails).

	_, err := u.handleCreateMessage(context.Background(), req)
	assert.Error(t, err)
	// If bug is present, the error message will be "no downstream session found for upstream session"
	// If bug is fixed, the error message should be "no session associated with request"

	assert.Equal(t, "no session associated with request", err.Error())
}
