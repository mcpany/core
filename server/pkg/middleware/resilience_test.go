// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestResilienceMiddleware_Execute_CircuitBreaker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock tool manager
	mockTM := tool.NewMockManagerInterface(ctrl)
	mw := NewResilienceMiddleware(mockTM)

	serviceID := "test-service"
	toolName := "test-tool"

	// Mock tool (Manual mock because generated one seems problematic or not used here)
	mockTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String(toolName),
				ServiceId: proto.String(serviceID),
			}.Build()
		},
	}

	// Configure resilience
	cbConfig := configv1.CircuitBreakerConfig_builder{
		FailureRateThreshold: proto.Float64(0.5),
		ConsecutiveFailures:  proto.Int32(2),
		OpenDuration:         durationpb.New(100 * time.Millisecond),
		HalfOpenRequests:     proto.Int32(1),
	}.Build()

	resilienceConfig := configv1.ResilienceConfig_builder{
		CircuitBreaker: cbConfig,
	}.Build()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		Resilience: resilienceConfig,
	}.Build()

	serviceInfo := &tool.ServiceInfo{
		Name:   serviceID,
		Config: svcConfig,
	}

	mockTM.EXPECT().GetTool(toolName).Return(mockTool, true).AnyTimes()
	mockTM.EXPECT().GetServiceInfo(serviceID).Return(serviceInfo, true).AnyTimes()

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: toolName}

	// 1. First failure
	err1 := errors.New("execution failed")
	next1 := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return nil, err1
	}
	_, gotErr1 := mw.Execute(ctx, req, next1)
	assert.Equal(t, err1, gotErr1)

	// 2. Second failure (Trips breaker)
	next2 := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return nil, err1
	}
	_, gotErr2 := mw.Execute(ctx, req, next2)
	assert.Equal(t, err1, gotErr2)

	// 3. Third call (Should be blocked by Circuit Breaker)
	called3 := false
	next3 := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		called3 = true
		return "success", nil
	}
	_, gotErr3 := mw.Execute(ctx, req, next3)
	assert.IsType(t, &resilience.CircuitBreakerOpenError{}, gotErr3)
	assert.False(t, called3, "Next should not be called when breaker is open")

	// 4. Wait for open duration
	time.Sleep(600 * time.Millisecond)

	// 5. Fourth call (Half-Open -> Success closes breaker)
	called4 := false
	next4 := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		called4 = true
		return "recovered", nil
	}
	res4, gotErr4 := mw.Execute(ctx, req, next4)
	assert.NoError(t, gotErr4)
	assert.Equal(t, "recovered", res4)
	assert.True(t, called4)
}

func TestResilienceMiddleware_Execute_Retry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := tool.NewMockManagerInterface(ctrl)
	mw := NewResilienceMiddleware(mockTM)

	serviceID := "retry-service"
	toolName := "retry-tool"

	mockTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String(toolName),
				ServiceId: proto.String(serviceID),
			}.Build()
		},
	}

	retryConfig := configv1.RetryConfig_builder{
		NumberOfRetries: proto.Int32(2),
		BaseBackoff:     durationpb.New(1 * time.Millisecond),
		MaxBackoff:      durationpb.New(10 * time.Millisecond),
	}.Build()

	resilienceConfig := configv1.ResilienceConfig_builder{
		RetryPolicy: retryConfig,
	}.Build()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		Resilience: resilienceConfig,
	}.Build()

	serviceInfo := &tool.ServiceInfo{
		Name:   serviceID,
		Config: svcConfig,
	}

	mockTM.EXPECT().GetTool(toolName).Return(mockTool, true).AnyTimes()
	mockTM.EXPECT().GetServiceInfo(serviceID).Return(serviceInfo, true).AnyTimes()

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: toolName}

	attempts := 0
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		attempts++
		if attempts <= 2 {
			return nil, errors.New("transient error")
		}
		return "success", nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.Equal(t, 3, attempts, "Should have run 3 times (1 initial + 2 retries)")
}

func TestResilienceMiddleware_NoConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := tool.NewMockManagerInterface(ctrl)
	mw := NewResilienceMiddleware(mockTM)

	serviceID := "no-config-service"
	toolName := "tool"

	mockTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String(toolName),
				ServiceId: proto.String(serviceID),
			}.Build()
		},
	}

	// No resilience config
	serviceInfo := &tool.ServiceInfo{
		Name:   serviceID,
		Config: &configv1.UpstreamServiceConfig{},
	}

	mockTM.EXPECT().GetTool(toolName).Return(mockTool, true).AnyTimes()
	mockTM.EXPECT().GetServiceInfo(serviceID).Return(serviceInfo, true).AnyTimes()

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: toolName}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "direct", nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "direct", res)
}
