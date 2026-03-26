// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
)

type mockToolManager struct {
	tool.ManagerInterface
	tools map[string]tool.Tool
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.tools[t.Tool().GetName()] = t
	return nil
}

// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t, ok := m.tools[toolName]
	return t, ok
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// AddServiceInfo ...
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

type mockPromptManager struct {
	prompt.ManagerInterface
	prompts map[string]prompt.Prompt
}

func newMockPromptManager() *mockPromptManager {
	return &mockPromptManager{
		prompts: make(map[string]prompt.Prompt),
	}
}

// AddPrompt ...
// Summary: AddPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.prompts[p.Prompt().Name] = p
}

// GetPrompt ...
// Summary: GetPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	p, ok := m.prompts[name]
	return p, ok
}

type mockResourceManager struct {
	resource.ManagerInterface
	resources map[string]resource.Resource
}

func newMockResourceManager() *mockResourceManager {
	return &mockResourceManager{
		resources: make(map[string]resource.Resource),
	}
}

// AddResource ...
// Summary: AddResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.resources[r.Resource().URI] = r
}

// GetResource ...
// Summary: GetResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	r, ok := m.resources[uri]
	return r, ok
}

// OnListChanged ...
// Summary: OnListChanged
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// Subscribe ...
// Summary: Subscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

type mockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

// Authenticate ...
// Summary: Authenticate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

// RoundTrip ...
// Summary: RoundTrip
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}
