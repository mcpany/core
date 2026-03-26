// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockHealthCheckerUpstream implements upstream.Upstream and upstream.HealthChecker
// MockHealthCheckerUpstream implements upstream.Upstream and upstream.HealthChecker
// Summary: MockHealthCheckerUpstream
	upstream.Upstream
	healthErr error
}

// CheckHealth ...
// Summary: CheckHealth
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.healthErr
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
	serviceID, _ := util.SanitizeServiceName(serviceConfig.GetName())
	return serviceID, nil, nil, nil
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

// TestErrorPropagation_HealthCheckFailure ...
// Summary: TestErrorPropagation_HealthCheckFailure
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockU := &MockHealthCheckerUpstream{
		healthErr: errors.New("connection refused"),
	}

	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return mockU, nil
		},
	}

	// Use threadSafeToolManager from registry_test.go
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	serviceConfig.SetName("unhealthy-service")

	// Register service
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Verify error via GetServiceError
	msg, ok := registry.GetServiceError(serviceID)
	assert.True(t, ok)
	assert.Equal(t, "connection refused", msg)

	// Verify GetAllServices returns LastError
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, "connection refused", services[0].GetLastError())

	// Verify GetServiceConfig returns LastError
	cfg, ok := registry.GetServiceConfig(serviceID)
	assert.True(t, ok)
	assert.Equal(t, "connection refused", cfg.GetLastError())
}

// TestToolCountPropagation ...
// Summary: TestToolCountPropagation
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockU := &MockHealthCheckerUpstream{}
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return mockU, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	serviceConfig.SetName("tool-service")
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Add 3 tools for this service
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("t1"), ServiceId: proto.String(serviceID)}.Build()})
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("t2"), ServiceId: proto.String(serviceID)}.Build()})
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("t3"), ServiceId: proto.String(serviceID)}.Build()})

	// Add tool for another service
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("other"), ServiceId: proto.String("other")}.Build()})

	// Verify GetAllServices returns correct ToolCount
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, int32(3), services[0].GetToolCount())
}
