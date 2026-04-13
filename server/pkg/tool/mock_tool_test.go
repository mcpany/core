package tool

import (
	"context"
	"testing"
	"errors"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMockTool(t *testing.T) {
	mt := &MockTool{
		ToolFunc: func() *v1.Tool {
			t := &v1.Tool{}
			t.Name = "mock"
			return t
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{Name: "mock_mcp"}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return "executed", nil
		},
		GetCacheConfigFunc: func() *configv1.CacheConfig {
			return &configv1.CacheConfig{TtlSeconds: 60}
		},
	}

	assert.Equal(t, "mock", mt.Tool().Name)
	assert.Equal(t, "mock_mcp", mt.MCPTool().Name)
	assert.False(t, mt.IsStreaming())
	assert.Equal(t, int32(60), mt.GetCacheConfig().TtlSeconds)

	res, err := mt.Execute(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "executed", res)

	ch, err := mt.StreamExecute(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "executed", <-ch)
}

func TestMockTool_Defaults(t *testing.T) {
	mt := &MockTool{}

	assert.NotNil(t, mt.Tool())
	assert.Nil(t, mt.MCPTool())
	assert.False(t, mt.IsStreaming())
	assert.Nil(t, mt.GetCacheConfig())

	res, err := mt.Execute(context.Background(), nil)
	assert.NoError(t, err)
	assert.Nil(t, res)

	ch, err := mt.StreamExecute(context.Background(), nil)
	assert.NoError(t, err)
	assert.Nil(t, <-ch)
}

func TestMockTool_StreamError(t *testing.T) {
    mt := &MockTool{
        ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
            return nil, errors.New("mock error")
        },
    }

    ch, err := mt.StreamExecute(context.Background(), nil)
    assert.NoError(t, err)
    errRes := <-ch
    assert.Equal(t, errors.New("mock error"), errRes)
}
