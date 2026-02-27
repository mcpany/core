// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Mock tool for testing DynamicResource
// We implement tool.Tool interface explicitly to avoid field/method conflict with embedding
type mockTool struct {
	executeFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
	toolDef     *mcp_routerv1.Tool
}

// Ensure mockTool implements tool.Tool
var _ tool.Tool = (*mockTool)(nil)

func (m *mockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return "default content", nil
}

func (m *mockTool) Tool() *mcp_routerv1.Tool {
	if m.toolDef != nil {
		return m.toolDef
	}
	return mcp_routerv1.Tool_builder{ServiceId: proto.String("mock-service")}.Build()
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockTool) MCPTool() *mcp.Tool {
	// Minimal mock implementation
	return &mcp.Tool{
		Name: m.Tool().GetName(),
	}
}

func TestNewDynamicResource(t *testing.T) {
	def := configv1.ResourceDefinition_builder{
		Uri:         proto.String("test://resource"),
		Name:        proto.String("test-resource"),
		Title:       proto.String("Test Resource"),
		Description: proto.String("A test resource"),
		MimeType:    proto.String("text/plain"),
	}.Build()
	mockT := &mockTool{}

	dr, err := NewDynamicResource(def, mockT)
	require.NoError(t, err)
	assert.Equal(t, "test://resource", dr.Resource().URI)
	assert.Equal(t, "test-resource", dr.Resource().Name)
	assert.Equal(t, "mock-service", dr.Service())

	_, err = NewDynamicResource(nil, mockT)
	assert.Error(t, err)

	_, err = NewDynamicResource(def, nil)
	assert.Error(t, err)
}

func TestDynamicResource_Read_Types(t *testing.T) {
	def := configv1.ResourceDefinition_builder{
		Uri:      proto.String("test://resource"),
		MimeType: proto.String("text/plain"),
	}.Build()

	tests := []struct {
		name           string
		toolResult     any
		expectedText   string
		expectedBlob   []byte
		expectedError  string
	}{
		{
			name:         "String result",
			toolResult:   "hello world",
			expectedText: "hello world",
		},
		{
			name:         "Byte slice result",
			toolResult:   []byte("hello bytes"),
			expectedBlob: []byte("hello bytes"),
		},
		{
			name:         "Map result (JSON)",
			toolResult:   map[string]interface{}{"key": "value"},
			expectedText: `{"key":"value"}`,
		},
		{
			name:          "Unsupported type",
			toolResult:    12345,
			expectedError: "unsupported tool result type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &mockTool{
				executeFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
					return tt.toolResult, nil
				},
				toolDef: mcp_routerv1.Tool_builder{ServiceId: proto.String("svc")}.Build(),
			}
			dr, _ := NewDynamicResource(def, mt)

			res, err := dr.Read(context.Background())
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Len(t, res.Contents, 1)
				if tt.expectedText != "" {
					assert.Equal(t, tt.expectedText, res.Contents[0].Text)
				}
				if tt.expectedBlob != nil {
					assert.Equal(t, tt.expectedBlob, res.Contents[0].Blob)
				}
				assert.Equal(t, "text/plain", res.Contents[0].MIMEType)
			}
		})
	}
}

func TestDynamicResource_ToolFailure(t *testing.T) {
	def := configv1.ResourceDefinition_builder{Uri: proto.String("test://fail")}.Build()
	mt := &mockTool{
		executeFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, errors.New("execution boom")
		},
		toolDef: mcp_routerv1.Tool_builder{ServiceId: proto.String("svc")}.Build(),
	}
	dr, _ := NewDynamicResource(def, mt)

	_, err := dr.Read(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute tool")
	assert.Contains(t, err.Error(), "execution boom")
}

func TestDynamicResource_Subscribe(t *testing.T) {
	def := configv1.ResourceDefinition_builder{Uri: proto.String("test://sub")}.Build()
	mt := &mockTool{toolDef: mcp_routerv1.Tool_builder{ServiceId: proto.String("svc")}.Build()}
	dr, _ := NewDynamicResource(def, mt)

	err := dr.Subscribe(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

// Integration-like test ensuring protobuf generated code works with the resource
func TestDynamicResource_ProtoIntegration(t *testing.T) {
	// Create a real proto tool definition
	toolProto := mcp_routerv1.Tool_builder{
		Name:      proto.String("real-tool"),
		ServiceId: proto.String("real-service"),
	}.Build()

	mt := &mockTool{
		toolDef: toolProto,
		executeFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "real content", nil
		},
	}

	def := configv1.ResourceDefinition_builder{
		Uri:  proto.String("real://resource"),
		Name: proto.String("Real Resource"),
	}.Build()

	dr, err := NewDynamicResource(def, mt)
	require.NoError(t, err)

	assert.Equal(t, "real-service", dr.Service())

	res, err := dr.Read(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "real content", res.Contents[0].Text)
}
