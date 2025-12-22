package middleware_test

import (
	"context"
	"encoding/json"
	"testing"

	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

type rateLimitMockTool struct {
	toolProto *v1.Tool
	mock.Mock
}

func (m *rateLimitMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *rateLimitMockTool) Tool() *v1.Tool {
	return m.toolProto
}

func (m *rateLimitMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *rateLimitMockTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.toolProto)
	return t
}

type rateLimitMockToolManager struct {
	tool.ManagerInterface
	mock.Mock
}

func (m *rateLimitMockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *rateLimitMockToolManager) GetTool(toolName string) (tool.Tool, bool) {
	args := m.Called(toolName)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

func (m *rateLimitMockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func TestRateLimitMiddleware(t *testing.T) {
	const successResult = "success"
	t.Run("rate limit allowed", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(1),
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(1),
			Burst:             proto.Int64(1),
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		// First request should succeed
		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)

		// Immediate second request should fail
		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("rate limit config update", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		// Initial Config: 10 RPS, Burst 10
		rlConfig1 := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(10),
		}.Build()

		serviceInfo1 := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig1,
			}.Build(),
		}

		// Updated Config: 1 RPS, Burst 1
		rlConfig2 := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(1),
			Burst:             proto.Int64(1),
		}.Build()

		serviceInfo2 := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig2,
			}.Build(),
		}

		// Mock sequence:
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo1, true).Once()
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo2, true).Twice()

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		// 1. Allowed (Config 1: 10 RPS)
		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)

		// 2. Allowed (Config 2: 1 RPS, Burst 1)
		// Since we updated burst to 1, and we have tokens (from 10), but burst cap reduces tokens to 1.
		// Wait, if I had 9 tokens. New Burst 1. Tokens -> 1.
		// Allow() consumes 1. Tokens -> 0. Success.
		result, err = rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)

		// 3. Blocked (Config 2: 1 RPS, Burst 1, Tokens 0)
		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("rate limit disabled", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled: proto.Bool(false),
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("service info not found", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(nil, false)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("redis rate limit allowed", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
			return db
		})
		defer middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
			return redis.NewClient(opts)
		})

		fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		middleware.SetTimeNowForTests(func() time.Time {
			return fixedTime
		})
		defer middleware.SetTimeNowForTests(time.Now)

		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		storageRedis := configv1.RateLimitConfig_STORAGE_REDIS
		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(10),
			Storage:           &storageRedis,
			Redis: busproto.RedisBus_builder{
				Address: proto.String("localhost:6379"),
			}.Build(),
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		// Mock Redis calls
		// The script returns 1 (allowed)
		mockRedis.ExpectEval(
			middleware.RedisRateLimitScript,
			[]string{"ratelimit:service"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetVal(int64(1))

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})
}
