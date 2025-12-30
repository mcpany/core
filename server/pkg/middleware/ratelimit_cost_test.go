// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/tokenizer"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	github_com_modelcontextprotocol_go_sdk_mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// Define Mocks locally to avoid dependency issues with main package tests if necessary

// MockToolManagerForCost is a mock for tool.ManagerInterface
type MockToolManagerForCost struct {
	mock.Mock
}

func (m *MockToolManagerForCost) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

func (m *MockToolManagerForCost) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *MockToolManagerForCost) ListTools() []tool.Tool { return nil }
func (m *MockToolManagerForCost) AddTool(t tool.Tool) error { return nil }
func (m *MockToolManagerForCost) AddServiceInfo(id string, info *tool.ServiceInfo) {}
func (m *MockToolManagerForCost) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) { return nil, nil }
func (m *MockToolManagerForCost) SetMCPServer(s tool.MCPServerProvider) {}
func (m *MockToolManagerForCost) ClearToolsForService(serviceKey string) {}
func (m *MockToolManagerForCost) SetProfiles(profiles []string, definitions []*configv1.ProfileDefinition) {}
func (m *MockToolManagerForCost) AddMiddleware(middleware tool.ExecutionMiddleware) {}
func (m *MockToolManagerForCost) ListServices() []*tool.ServiceInfo { return nil }

// MockToolForCost is a mock for tool.Tool
type MockToolForCost struct {
	mock.Mock
}

func (m *MockToolForCost) Tool() *v1.Tool {
	args := m.Called()
	return args.Get(0).(*v1.Tool)
}

func (m *MockToolForCost) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *MockToolForCost) Service() string {
	return "service1"
}
// Return concrete type instead of interface proxy
func (m *MockToolForCost) MCPTool() *github_com_modelcontextprotocol_go_sdk_mcp.Tool { return nil }
func (m *MockToolForCost) GetCacheConfig() *configv1.CacheConfig { return nil }

// Need dummy type for MCPTool return to satisfy interface
type github_com_modelcontextprotocol_go_sdk_mcp_Tool struct{}


func TestRateLimitMiddleware_EstimateTokenCost(t *testing.T) {
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
			expected: 1, // 5 chars / 4 = 1
		},
		{
			name: "long string",
			inputs: map[string]any{
				"arg1": "this is a longer string that should cost more tokens",
			},
			expected: 13, // 52 chars / 4 = 13
		},
		{
			name: "multiple args",
			inputs: map[string]any{
				"arg1": "hello",
				"arg2": "world",
			},
			expected: 2, // 10 chars / 4 = 2.5 -> 2
		},
		{
			name: "non-string args",
			inputs: map[string]any{
				"arg1": 12345, // "12345" -> 5 chars
			},
			expected: 1, // 5 / 4 = 1
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

func TestRateLimitMiddleware_EstimateTokenCost_WordTokenizer(t *testing.T) {
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
				"arg1": "hello world", // 2 words * 1.3 = 2.6 -> 2
			},
			expected: 2,
		},
		{
			name: "sentence",
			inputs: map[string]any{
				"arg1": "this is a test sentence", // 5 words * 1.3 = 6.5 -> 6
			},
			expected: 6,
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

func TestRateLimitMiddleware_AllowN(t *testing.T) {
	mockManager := new(MockToolManagerForCost)
	middleware := NewRateLimitMiddleware(mockManager)

	serviceID := "test-service"
	toolName := "test-tool"

	// Mock Tool
	toolDef := &v1.Tool{}
	toolDef.ServiceId = proto.String(serviceID)

	mockTool := new(MockToolForCost)
	mockTool.On("Tool").Return(toolDef)
	mockManager.On("GetTool", toolName).Return(mockTool, true)

	// Config with Token Cost Metric
	config := &configv1.RateLimitConfig{}
	config.IsEnabled = proto.Bool(true)
	config.RequestsPerSecond = proto.Float64(10)
	config.Burst = proto.Int64(20)
	config.CostMetric = configv1.RateLimitConfig_COST_METRIC_TOKENS.Enum()
	config.Storage = configv1.RateLimitConfig_STORAGE_MEMORY.Enum()

	mockManager.On("GetServiceInfo", serviceID).Return(&tool.ServiceInfo{
		Name:   "Test Service",
		Config: &configv1.UpstreamServiceConfig{
			RateLimit: config,
		},
	}, true)

	ctx := context.Background()

	// 1. Request with low cost (allowed)
	// Cost ~2 tokens
	args1 := map[string]any{
		"arg": "small input",
	}
	req1 := &tool.ExecutionRequest{
		ToolName: toolName,
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
	for i := range longString { longString[i] = 'a' }

	args2 := map[string]any{
		"arg": string(longString),
	}
	req2 := &tool.ExecutionRequest{
		ToolName: toolName,
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
