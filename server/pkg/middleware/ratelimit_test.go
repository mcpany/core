// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
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

func (m *rateLimitMockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	return true
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
			IsEnabled:         true,
			RequestsPerSecond: 10,
			Burst:             1,
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
			IsEnabled:         true,
			RequestsPerSecond: 1,
			Burst:             1,
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
			IsEnabled:         true,
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()

		serviceInfo1 := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig1,
			}.Build(),
		}

		// Updated Config: 1 RPS, Burst 1
		rlConfig2 := configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 1,
			Burst:             1,
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
			IsEnabled: false,
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
		middleware.SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
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
			IsEnabled:         true,
			RequestsPerSecond: 10,
			Burst:             10,
			Storage:           storageRedis,
			Redis: busproto.RedisBus_builder{
				Address: proto.String("127.0.0.1:6379"),
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
		s := redis.NewScript(middleware.RedisRateLimitScript)
		// Updated expectation: key now includes "service" scope suffix "service:service"
		mockRedis.ExpectEvalSha(
			s.Hash(),
			[]string{"ratelimit:service:service"},
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

	t.Run("redis rate limit error", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		middleware.SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
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
			IsEnabled:         true,
			RequestsPerSecond: 10,
			Burst:             10,
			Storage:           storageRedis,
			Redis: busproto.RedisBus_builder{
				Address: proto.String("127.0.0.1:6379"),
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
		s := redis.NewScript(middleware.RedisRateLimitScript)
		mockRedis.ExpectEvalSha(
			s.Hash(),
			[]string{"ratelimit:service:service"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetErr(errors.New("redis connection failed"))

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit check failed")
		assert.Nil(t, result)
		assert.False(t, nextCalled)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestRateLimitMiddleware_Partitioning(t *testing.T) {
	const successResult = "success"

	setupMiddleware := func(keyBy configv1.RateLimitConfig_KeyBy) (*middleware.RateLimitMiddleware, *rateLimitMockToolManager, *rateLimitMockTool) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 1,
			Burst:             1,
			KeyBy:             keyBy,
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		return rlMiddleware, mockToolManager, mockTool
	}

	t.Run("partition by ip", func(t *testing.T) {
		rlMiddleware, _, mockTool := setupMiddleware(configv1.RateLimitConfig_KEY_BY_IP)

		execute := func(ip string) error {
			req := &tool.ExecutionRequest{
				ToolName:   "service.test-tool",
				ToolInputs: json.RawMessage(`{}`),
			}
			ctx := context.Background()
			if ip != "" {
				ctx = util.ContextWithRemoteIP(ctx, ip)
			}
			ctx = tool.NewContextWithTool(ctx, mockTool)

			next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				return successResult, nil
			}
			_, err := rlMiddleware.Execute(ctx, req, next)
			return err
		}

		// IP 1: Request 1 -> Allow
		assert.NoError(t, execute("1.1.1.1"))

		// IP 1: Request 2 -> Block (1 RPS exceeded)
		assert.Error(t, execute("1.1.1.1"))

		// IP 2: Request 1 -> Allow (Different bucket)
		assert.NoError(t, execute("2.2.2.2"))
	})

	t.Run("partition by user", func(t *testing.T) {
		rlMiddleware, _, mockTool := setupMiddleware(configv1.RateLimitConfig_KEY_BY_USER_ID)

		execute := func(user string) error {
			req := &tool.ExecutionRequest{
				ToolName:   "service.test-tool",
				ToolInputs: json.RawMessage(`{}`),
			}
			ctx := context.Background()
			if user != "" {
				ctx = auth.ContextWithUser(ctx, user)
			}
			ctx = tool.NewContextWithTool(ctx, mockTool)

			next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				return successResult, nil
			}
			_, err := rlMiddleware.Execute(ctx, req, next)
			return err
		}

		// User A: Request 1 -> Allow
		assert.NoError(t, execute("userA"))

		// User A: Request 2 -> Block
		assert.Error(t, execute("userA"))

		// User B: Request 1 -> Allow
		assert.NoError(t, execute("userB"))
	})

	t.Run("partition by api key header", func(t *testing.T) {
		rlMiddleware, _, mockTool := setupMiddleware(configv1.RateLimitConfig_KEY_BY_API_KEY)

		execute := func(apiKey string) error {
			req := &tool.ExecutionRequest{
				ToolName:   "service.test-tool",
				ToolInputs: json.RawMessage(`{}`),
			}
			ctx := context.Background()

			// Inject HTTP request with header
			httpReq, _ := http.NewRequest("POST", "/", nil)
			if apiKey != "" {
				httpReq.Header.Set("X-API-Key", apiKey)
			}
			ctx = context.WithValue(ctx, middleware.HTTPRequestContextKey, httpReq)

			ctx = tool.NewContextWithTool(ctx, mockTool)

			next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				return successResult, nil
			}
			_, err := rlMiddleware.Execute(ctx, req, next)
			return err
		}

		// Key A: Request 1 -> Allow
		assert.NoError(t, execute("keyA"))

		// Key A: Request 2 -> Block
		assert.Error(t, execute("keyA"))

		// Key B: Request 1 -> Allow
		assert.NoError(t, execute("keyB"))
	})
}
