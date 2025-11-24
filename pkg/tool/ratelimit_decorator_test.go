package tool

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/policy/ratelimit"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Ptr[T any](v T) *T {
	return &v
}

// mockTool is a mock implementation of the Tool interface for testing.
type mockTool struct {
	tool          *v1.Tool
	executionFunc func(ctx context.Context, req *ExecutionRequest) (any, error)
}

func (m *mockTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.executionFunc != nil {
		return m.executionFunc(ctx, req)
	}
	return "success", nil
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestRateLimitedTool(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		tool: &v1.Tool{Name: Ptr("test-tool")},
	}

	// Create a rate limiter that allows 1 request per second with a burst of 1
	limiter := ratelimit.NewInMemoryLimiter(1, 1)

	// Wrap the mock tool with the rate limiter
	rateLimitedTool := NewRateLimitedTool(mt, limiter)

	// First request should be allowed
	_, err := rateLimitedTool.Execute(context.Background(), &ExecutionRequest{})
	assert.NoError(t, err)

	// Second request immediately after should be rate limited
	_, err = rateLimitedTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.ResourceExhausted, st.Code())

	// Wait for the rate limiter to allow another request
	time.Sleep(1 * time.Second)

	// Third request should be allowed
	_, err = rateLimitedTool.Execute(context.Background(), &ExecutionRequest{})
	assert.NoError(t, err)
}
