// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockToolManager for testing
// MockToolManager for testing
// Summary: MockToolManager
	mock.Mock
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
	args := m.Called(name)
	if t := args.Get(0); t != nil {
		return t.(tool.Tool), args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(id)
	if s := args.Get(0); s != nil {
		return s.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

// Implement other interface methods as no-ops or panics if needed
// Implement other interface methods as no-ops or panics if needed
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ListMCPTools ...
// Summary: ListMCPTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// IsServiceAllowed ...
// Summary: IsServiceAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// AddMiddleware ...
// Summary: AddMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ToolMatchesProfile ...
// Summary: ToolMatchesProfile
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

// MockTool ...
// Summary: MockTool
	name      string
	serviceID string
}

// Tool ...
// Summary: Tool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return v1.Tool_builder{Name: proto.String(t.name), ServiceId: proto.String(t.serviceID)}.Build()
}
// MCPTool ...
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Execute ...
// Summary: Execute
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
// GetCacheConfig ...
// Summary: GetCacheConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// TestRateLimitMiddleware_Granular ...
// Summary: TestRateLimitMiddleware_Granular
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	tm := &MockToolManager{}
	mw := middleware.NewRateLimitMiddleware(tm)

	serviceID := "test-service"
	toolName := "restricted-tool"
	otherToolName := "normal-tool"

	// Setup Config
	// Service Limit: 100 RPS
	// Tool Limit: 1 RPS (Burst 1)
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		RateLimit: configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 100.0,
			Burst:             100,
			ToolLimits: map[string]*configv1.RateLimitConfig{
				toolName: configv1.RateLimitConfig_builder{
					IsEnabled:         true,
					RequestsPerSecond: 1.0,
					Burst:             1,
				}.Build(),
			},
		}.Build(),
	}.Build()

	tm.On("GetTool", toolName).Return(&MockTool{name: toolName, serviceID: serviceID}, true)
	tm.On("GetTool", otherToolName).Return(&MockTool{name: otherToolName, serviceID: serviceID}, true)
	tm.On("GetServiceInfo", serviceID).Return(&tool.ServiceInfo{Name: serviceID, Config: config}, true)

	ctx := context.Background()
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	// Test Case 1: Restricted Tool
	// 1st call should succeed
	req1 := &tool.ExecutionRequest{ToolName: toolName}
	res, err := mw.Execute(ctx, req1, next)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)

	// 2nd call immediate should fail (Burst 1)
	res, err = mw.Execute(ctx, req1, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded for tool")

	// Test Case 2: Normal Tool (falls back to service limit)
	// Should pass easily
	req2 := &tool.ExecutionRequest{ToolName: otherToolName}
	res, err = mw.Execute(ctx, req2, next)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}

// TestRateLimitMiddleware_ServiceLimitFallback ...
// Summary: TestRateLimitMiddleware_ServiceLimitFallback
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	tm := &MockToolManager{}
	mw := middleware.NewRateLimitMiddleware(tm)

	serviceID := "test-service-fallback"
	toolName := "normal-tool"

	// Service Limit: 1 RPS
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		RateLimit: configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 1.0,
			Burst:             1,
		}.Build(),
	}.Build()

	tm.On("GetTool", toolName).Return(&MockTool{name: toolName, serviceID: serviceID}, true)
	tm.On("GetServiceInfo", serviceID).Return(&tool.ServiceInfo{Name: serviceID, Config: config}, true)

	ctx := context.Background()
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	req := &tool.ExecutionRequest{ToolName: toolName}

	// 1st call ok
	_, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)

	// 2nd call blocked
	_, err = mw.Execute(ctx, req, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded for service")
}

// GetAllowedServiceIDs ...
// Summary: GetAllowedServiceIDs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

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
	return 0
}
