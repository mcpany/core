// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	github_com_modelcontextprotocol_go_sdk_mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// Define Mocks locally to avoid dependency issues with main package tests if necessary

// MockToolManagerForCost is a mock for tool.ManagerInterface
// MockToolManagerForCost is a mock for tool.ManagerInterface
// Summary: MockToolManagerForCost
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
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
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
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

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
	return nil
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
}
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
	return true
}

// MockToolForCost is a mock for tool.Tool
// MockToolForCost is a mock for tool.Tool
// Summary: MockToolForCost
	mock.Mock
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
	args := m.Called()
	return args.Get(0).(*v1.Tool)
}

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
	return "service1"
}

// Return concrete type instead of interface proxy
// Return concrete type instead of interface proxy
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// Need dummy type for MCPTool return to satisfy interface
type github_com_modelcontextprotocol_go_sdk_mcp_Tool struct{}

// TestRateLimitMiddleware_EstimateTokenCost ...
// Summary: TestRateLimitMiddleware_EstimateTokenCost
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Default SimpleTokenizer (4 chars/token)
	m := NewRateLimitMiddleware(&MockToolManagerForCost{})

	tests := []struct {
		name     string
		inputs   map[string]any
		expected int
	}{
		{
			name:     "empty inputs",
			inputs:   map[string]any{},
			expected: 1, // Minimum cost
		},
		{
			name: "short string",
			inputs: map[string]any{
				"arg1": "hello",
			},
			expected: 2, // key "arg1" (1) + val "hello" (1) = 2
		},
		{
			name: "long string",
			inputs: map[string]any{
				"arg1": "this is a longer string that should cost more tokens",
			},
			expected: 14, // key "arg1" (1) + val (13) = 14
		},
		{
			name: "multiple args",
			inputs: map[string]any{
				"arg1": "hello",
				"arg2": "world",
			},
			expected: 4, // "arg1"(1)+"hello"(1) + "arg2"(1)+"world"(1) = 4
		},
		{
			name: "non-string args",
			inputs: map[string]any{
				"arg1": 12345, // "12345" -> 5 chars
			},
			expected: 2, // key "arg1"(1) + val (1) = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &tool.ExecutionRequest{
				ToolInputs: func() json.RawMessage {
					b, _ := json.Marshal(tt.inputs)
					return b
				}(),
				Arguments: tt.inputs,
			}
			cost := m.estimateTokenCost(req)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

// TestRateLimitMiddleware_EstimateTokenCost_WordTokenizer ...
// Summary: TestRateLimitMiddleware_EstimateTokenCost_WordTokenizer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// WordTokenizer (1.3 * words)
	wt := tokenizer.NewWordTokenizer()
	m := NewRateLimitMiddleware(&MockToolManagerForCost{}, WithTokenizer(wt))

	tests := []struct {
		name     string
		inputs   map[string]any
		expected int
	}{
		{
			name: "hello world",
			inputs: map[string]any{
				"arg1": "hello world", // key "arg1" (1) + val (2) = 3
			},
			expected: 3,
		},
		{
			name: "sentence",
			inputs: map[string]any{
				"arg1": "this is a test sentence", // key "arg1" (1) + val (6) = 7
			},
			expected: 7,
		},
		{
			name: "slice int",
			inputs: map[string]any{
				"list": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			// key "list" (1) + val (10 items * 1.3 = 13) = 14
			// Prior to fix, this would be 1 + 10 = 11
			expected: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &tool.ExecutionRequest{
				ToolInputs: func() json.RawMessage {
					b, _ := json.Marshal(tt.inputs)
					return b
				}(),
				Arguments: tt.inputs,
			}
			cost := m.estimateTokenCost(req)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

// TestRateLimitMiddleware_AllowN ...
// Summary: TestRateLimitMiddleware_AllowN
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockManager := new(MockToolManagerForCost)
	middleware := NewRateLimitMiddleware(mockManager)

	serviceID := "test-service"
	toolName := "test-tool"

	// Mock Tool
	toolDef := v1.Tool_builder{
		ServiceId: proto.String(serviceID),
	}.Build()

	mockTool := new(MockToolForCost)
	mockTool.On("Tool").Return(toolDef)
	mockManager.On("GetTool", toolName).Return(mockTool, true)

	// Config with Token Cost Metric
	config := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10.0,
		Burst:             20,
		CostMetric:        configv1.RateLimitConfig_COST_METRIC_TOKENS,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
	}.Build()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		RateLimit: config,
	}.Build()

	mockManager.On("GetServiceInfo", serviceID).Return(&tool.ServiceInfo{
		Name:   "Test Service",
		Config: svcConfig,
	}, true)

	ctx := context.Background()

	// 1. Request with low cost (allowed)
	// Cost ~2 tokens
	args1 := map[string]any{
		"arg": "small input",
	}
	req1 := &tool.ExecutionRequest{
		ToolName:  toolName,
		Arguments: args1,
	}
	b1, _ := json.Marshal(args1)
	req1.ToolInputs = b1

	nextCalled := false
	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	// First call consumes ~2 tokens from 20 burst. Remaining: 18.
	_, err := middleware.Execute(ctx, req1, next)
	assert.NoError(t, err)
	assert.True(t, nextCalled)

	// 2. Request with high cost (exceeds remaining burst)
	// We want to consume > 18 tokens. 18 * 4 = 72 chars.
	// Let's use 100 chars -> 25 tokens.
	longString := make([]byte, 100)
	for i := range longString {
		longString[i] = 'a'
	}

	args2 := map[string]any{
		"arg": string(longString),
	}
	req2 := &tool.ExecutionRequest{
		ToolName:  toolName,
		Arguments: args2,
	}
	b2, _ := json.Marshal(args2)
	req2.ToolInputs = b2

	nextCalled = false
	// This should fail because 25 > 18 (remaining)
	_, err = middleware.Execute(ctx, req2, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	assert.False(t, nextCalled)
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
