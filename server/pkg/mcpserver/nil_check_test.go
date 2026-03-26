// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
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

// Prompt ...
// Summary: Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if !n.called {
		n.called = true
		return &mcp.Prompt{Name: "test-prompt"}
	}
	return nil
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Definition ...
// Summary: Definition
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

type nilResource struct {
	called bool
}

// Resource ...
// Summary: Resource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if !n.called {
		n.called = true
		return &mcp.Resource{Name: "test-resource", URI: "test://resource"}
	}
	return nil
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Read ...
// Summary: Read
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

// TestListPrompts_NilCheck ...
// Summary: TestListPrompts_NilCheck
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	s, err := NewServer(context.Background(), toolManager, promptManager, resourceManager, authManager, serviceRegistry, nil, busProvider, false)
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

// TestListResources_NilCheck ...
// Summary: TestListResources_NilCheck
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	s, err := NewServer(context.Background(), toolManager, promptManager, resourceManager, authManager, serviceRegistry, nil, busProvider, false)
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
