// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type hunterMockUpstream struct {
	upstream.Upstream
}

// Register ...
// Summary: Register
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return "mock-service-key", nil, nil, nil
}
// Shutdown ...
// Summary: Shutdown
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

type hunterMockFactory struct {
	factory.Factory
}

// NewUpstream ...
// Summary: NewUpstream
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &hunterMockUpstream{}, nil
}

type hunterMockToolManager struct {
	tool.ManagerInterface
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
// ClearToolsForService ...
// Summary: ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ExecuteTool ...
// Summary: ExecuteTool
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
// GetToolCountForService ...
// Summary: GetToolCountForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// TestServiceRegistry_SecretsLeak ...
// Summary: TestServiceRegistry_SecretsLeak
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &hunterMockFactory{}
	tm := &hunterMockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("secret-service")
	serviceConfig.SetAuthentication(configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			VerificationValue: proto.String("SUPER_SECRET_VALUE"),
		}.Build(),
	}.Build())

	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Get the config via GetServiceConfig
	retrievedConfig, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)

	auth := retrievedConfig.GetAuthentication()
	require.NotNil(t, auth)
	apiKey := auth.GetApiKey()
	require.NotNil(t, apiKey)

	// This should fail currently
	assert.Empty(t, apiKey.GetVerificationValue(), "API Key secret should be scrubbed in GetServiceConfig")

	// Get all services
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	require.Len(t, services, 1)

	authAll := services[0].GetAuthentication()
	require.NotNil(t, authAll)
	apiKeyAll := authAll.GetApiKey()
	// This should fail currently
	assert.Empty(t, apiKeyAll.GetVerificationValue(), "API Key secret should be scrubbed in GetAllServices")
}
