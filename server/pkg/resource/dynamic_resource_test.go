package resource

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockTool is a mock implementation of the tool.Tool interface.
type MockTool struct {
	mock.Mock
}

func (m *MockTool) Tool() *v1.Tool {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*v1.Tool)
}

func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*configv1.CacheConfig)
}

func (m *MockTool) MCPTool() *mcp.Tool {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*mcp.Tool)
}

func TestNewDynamicResource(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		def := configv1.ResourceDefinition_builder{
			Uri:         proto.String("test-uri"),
			Name:        proto.String("test-name"),
			Title:       proto.String("test-title"),
			Description: proto.String("test-description"),
			MimeType:    proto.String("text/plain"),
			Size:        proto.Int64(123),
		}.Build()
		mockTool := new(MockTool)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, err := NewDynamicResource(def, mockTool)
		require.NoError(t, err)
		assert.NotNil(t, dr)
		assert.Equal(t, "test-uri", dr.Resource().URI)
		assert.Equal(t, "test-name", dr.Resource().Name)
		assert.Equal(t, "test-title", dr.Resource().Title)
		assert.Equal(t, "test-description", dr.Resource().Description)
		assert.Equal(t, "text/plain", dr.Resource().MIMEType)
		assert.Equal(t, int64(123), dr.Resource().Size)
	})

	t.Run("nil definition", func(t *testing.T) {
		mockTool := new(MockTool)
		_, err := NewDynamicResource(nil, mockTool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resource definition is nil")
	})

	t.Run("nil tool", func(t *testing.T) {
		def := &configv1.ResourceDefinition{}
		_, err := NewDynamicResource(def, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tool is nil")
	})
}

func TestDynamicResource_Read(t *testing.T) {
	def := configv1.ResourceDefinition_builder{
		Uri: proto.String("test-uri"),
	}.Build()

	t.Run("string result", func(t *testing.T) {
		mockTool := new(MockTool)
		mockTool.On("Execute", mock.Anything, mock.Anything).Return("test-content", nil)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(def, mockTool)
		result, err := dr.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		assert.Equal(t, "test-uri", result.Contents[0].URI)
		assert.Equal(t, "test-content", result.Contents[0].Text)
	})

	t.Run("byte slice result", func(t *testing.T) {
		mockTool := new(MockTool)
		mockTool.On("Execute", mock.Anything, mock.Anything).Return([]byte("test-content"), nil)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(def, mockTool)
		result, err := dr.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		assert.Equal(t, "test-uri", result.Contents[0].URI)
		assert.Equal(t, []byte("test-content"), result.Contents[0].Blob)
	})

	t.Run("map result", func(t *testing.T) {
		mockTool := new(MockTool)
		mapResult := map[string]interface{}{"key": "value"}
		mockTool.On("Execute", mock.Anything, mock.Anything).Return(mapResult, nil)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(def, mockTool)
		result, err := dr.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		assert.Equal(t, "test-uri", result.Contents[0].URI)
		expectedJSON, _ := json.Marshal(mapResult)
		assert.JSONEq(t, string(expectedJSON), result.Contents[0].Text)
	})

	t.Run("unsupported result type", func(t *testing.T) {
		mockTool := new(MockTool)
		mockTool.On("Execute", mock.Anything, mock.Anything).Return(123, nil)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(def, mockTool)
		_, err := dr.Read(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported tool result type")
	})

	t.Run("execution error", func(t *testing.T) {
		mockTool := new(MockTool)
		mockTool.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("execution failed"))
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(def, mockTool)
		_, err := dr.Read(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute tool")
	})

	t.Run("json marshal error", func(t *testing.T) {
		mockTool := new(MockTool)
		// valid map but with a cyclic or unsupported value to trigger json.Marshal error
		// Using a channel which is not marshalable
		mapResult := map[string]interface{}{"key": make(chan int)}
		mockTool.On("Execute", mock.Anything, mock.Anything).Return(mapResult, nil)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(def, mockTool)
		_, err := dr.Read(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal tool result to JSON")
	})

	t.Run("MIMEType preservation", func(t *testing.T) {
		defWithMime := configv1.ResourceDefinition_builder{
			Uri:      proto.String("test-uri-mime"),
			MimeType: proto.String("application/json"),
		}.Build()

		mockTool := new(MockTool)
		mockTool.On("Execute", mock.Anything, mock.Anything).Return("{}", nil)
		mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
		dr, _ := NewDynamicResource(defWithMime, mockTool)
		result, err := dr.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		assert.Equal(t, "test-uri-mime", result.Contents[0].URI)
		assert.Equal(t, "{}", result.Contents[0].Text)
		assert.Equal(t, "application/json", result.Contents[0].MIMEType)
	})
}

func TestDynamicResource_Subscribe(t *testing.T) {
	def := configv1.ResourceDefinition_builder{
		Uri: proto.String("test-uri"),
	}.Build()
	mockTool := new(MockTool)
	mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
	dr, _ := NewDynamicResource(def, mockTool)

	err := dr.Subscribe(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscribing to dynamic resources is not yet implemented")
}

func TestDynamicResource_Service(t *testing.T) {
	def := configv1.ResourceDefinition_builder{
		Uri: proto.String("test-uri"),
	}.Build()
	mockTool := new(MockTool)
	mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
	dr, _ := NewDynamicResource(def, mockTool)

	assert.Equal(t, "test-service", dr.Service())
}
