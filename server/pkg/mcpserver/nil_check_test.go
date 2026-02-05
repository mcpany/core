// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nilPrompt struct {
	called bool
}

func (n *nilPrompt) Prompt() *mcp.Prompt {
	if !n.called {
		n.called = true
		return &mcp.Prompt{Name: "test-prompt"}
	}
	return nil
}

func (n *nilPrompt) Service() string { return "nil-service" }
func (n *nilPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

type nilResource struct {
	called bool
}

func (n *nilResource) Resource() *mcp.Resource {
	if !n.called {
		n.called = true
		return &mcp.Resource{Name: "test-resource", URI: "test://resource"}
	}
	return nil
}

func (n *nilResource) Service() string { return "nil-service" }
func (n *nilResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) { return nil, nil }
func (n *nilResource) Subscribe(ctx context.Context) error                       { return nil }

func TestListPrompts_NilCheck(t *testing.T) {
	// Setup dependencies
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	poolManager := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(poolManager, nil)
	serviceRegistry := serviceregistry.New(f, toolManager, promptManager, resourceManager, authManager)

	// Initialize server
	s, err := NewServer(context.Background(), toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	// Add a prompt that returns valid first time, then nil
	promptManager.AddPrompt(&nilPrompt{})

	// Call ListPrompts
	result, err := s.ListPrompts(context.Background(), &mcp.ListPromptsRequest{})
	require.NoError(t, err)

	// Verify that we do NOT get a nil prompt in the list
	foundNil := false
	for _, p := range result.Prompts {
		if p == nil {
			foundNil = true
			break
		}
	}
	assert.False(t, foundNil, "ListPrompts should NOT contain nil")
}

func TestListResources_NilCheck(t *testing.T) {
	// Setup dependencies
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	poolManager := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(poolManager, nil)
	serviceRegistry := serviceregistry.New(f, toolManager, promptManager, resourceManager, authManager)

	// Initialize server
	s, err := NewServer(context.Background(), toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	// Add a resource that returns valid first time, then nil
	resourceManager.AddResource(&nilResource{})

	// Call ListResources
	result, err := s.ListResources(context.Background(), &mcp.ListResourcesRequest{})
	require.NoError(t, err)

	// Verify that we do NOT get a nil resource in the list
	foundNil := false
	for _, r := range result.Resources {
		if r == nil {
			foundNil = true
			break
		}
	}
	assert.False(t, foundNil, "ListResources should NOT contain nil")
}
