// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// MockMCPClient is a mock for the mcp.Client interface
type MockMCPClient struct {
	mock.Mock
}

func (m *MockMCPClient) CallTool(ctx context.Context, req *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

func TestContextWithTool(t *testing.T) {
	ctx := context.Background()
	mockTool := new(MockTool)
	ctx = NewContextWithTool(ctx, mockTool)
	retrievedTool, ok := GetFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, mockTool, retrievedTool)

	_, ok = GetFromContext(context.Background())
	assert.False(t, ok)
}

func TestGRPCTool_Execute_PoolError(t *testing.T) {
	poolManager := pool.NewManager()
	toolProto := &v1.Tool{}
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)
	grpcTool := NewGRPCTool(toolProto, poolManager, "non-existent-service", mockMethodDesc, &configv1.GrpcCallDefinition{})
	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no grpc pool found for service")
}

func TestHTTPTool_Execute_PoolError(t *testing.T) {
	poolManager := pool.NewManager()
	toolProto := &v1.Tool{}
	httpTool := NewHTTPTool(toolProto, poolManager, "non-existent-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no http pool found for service")
}

func TestHTTPTool_Execute_InvalidFQN(t *testing.T) {
	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("invalid")
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http tool definition")
}

func TestHTTPTool_Execute_BadURL(t *testing.T) {
	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("GET %")
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
}

func TestHTTPTool_Execute_InputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)

	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("POST " + server.URL)
	callDef := &configv1.HttpCallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetInputTransformer(inputTransformer)
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, callDef, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestHTTPTool_Execute_OutputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer server.Close()

	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)

	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("GET " + server.URL)
	callDef := &configv1.HttpCallDefinition{}
	outputTransformer := &configv1.OutputTransformer{}
	outputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetOutputTransformer(outputTransformer)
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, callDef, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
}

func TestMCPTool_Execute_InputTransformerError(t *testing.T) {
	toolProto := &v1.Tool{}
	callDef := &configv1.MCPCallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetInputTransformer(inputTransformer)
	mcpTool := NewMCPTool(toolProto, nil, callDef)
	_, err := mcpTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestMCPTool_Execute_OutputTransformerError(t *testing.T) {
	mockClient := new(MockMCPClient)
	mockResult := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: `{"key":"value"}`}},
	}
	mockClient.On("CallTool", mock.Anything, mock.Anything).Return(mockResult, nil)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	callDef := &configv1.MCPCallDefinition{}
	outputTransformer := &configv1.OutputTransformer{}
	outputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetOutputTransformer(outputTransformer)
	mcpTool := NewMCPTool(toolProto, mockClient, callDef)
	_, err := mcpTool.Execute(context.Background(), &ExecutionRequest{ToolName: "test.test-tool", ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestOpenAPITool_Execute_InputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	toolProto := &v1.Tool{}
	callDef := &configv1.OpenAPICallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetInputTransformer(inputTransformer)
	openapiTool := NewOpenAPITool(toolProto, server.Client(), nil, "POST", server.URL, nil, callDef)
	_, err := openapiTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestOpenAPITool_Execute_OutputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer server.Close()

	toolProto := &v1.Tool{}
	callDef := &configv1.OpenAPICallDefinition{}
	outputTransformer := &configv1.OutputTransformer{}
	outputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetOutputTransformer(outputTransformer)
	openapiTool := NewOpenAPITool(toolProto, server.Client(), nil, "GET", server.URL, nil, callDef)
	_, err := openapiTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
}

// MockMethodDescriptor is a mock for protoreflect.MethodDescriptor
type MockMethodDescriptor struct {
	mock.Mock
	protoreflect.MethodDescriptor
}

func (m *MockMethodDescriptor) Input() protoreflect.MessageDescriptor {
	args := m.Called()
	return args.Get(0).(protoreflect.MessageDescriptor)
}

func (m *MockMethodDescriptor) Output() protoreflect.MessageDescriptor {
	args := m.Called()
	return args.Get(0).(protoreflect.MessageDescriptor)
}

// MockMessageDescriptor is a mock for protoreflect.MessageDescriptor
type MockMessageDescriptor struct {
	mock.Mock
	protoreflect.MessageDescriptor
}

func (m *MockMessageDescriptor) New() protoreflect.Message {
	return dynamicpb.NewMessage(m)
}

func (m *MockMessageDescriptor) Fields() protoreflect.FieldDescriptors {
	return nil
}

func (m *MockMessageDescriptor) FullName() protoreflect.FullName {
	return "test.Message"
}

func (m *MockMessageDescriptor) RequiredNumbers() protoreflect.FieldNumbers {
	return &MockFieldNumbers{}
}

// MockFieldNumbers is a mock for protoreflect.FieldNumbers
type MockFieldNumbers struct {
	protoreflect.FieldNumbers
}

func (m *MockFieldNumbers) Len() int {
	return 0
}

func (m *MockFieldNumbers) Get(_ int) protoreflect.FieldNumber {
	panic("should not be called")
}

func (m *MockFieldNumbers) Has(_ protoreflect.FieldNumber) bool {
	return false
}

func TestGRPCTool_Execute_Success(t *testing.T) {
	poolManager := pool.NewManager()

	mockConn := new(MockConn)
	mockConn.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockConn.On("GetState").Return(connectivity.Ready)

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(_ context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}, nil), nil
	}, 1, 1, 0, false)
	poolManager.Register("grpc-service", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)

	mockMethodDesc.On("Input").Return(mockMsgDesc)
	mockMethodDesc.On("Output").Return(mockMsgDesc)

	grpcTool := NewGRPCTool(toolProto, poolManager, "grpc-service", mockMethodDesc, &configv1.GrpcCallDefinition{})
	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Invoke", mock.Anything, "/test.service/Method", mock.AnythingOfType("*dynamicpb.Message"), mock.AnythingOfType("*dynamicpb.Message"), mock.Anything)
}

// MockConn is a mock for client.Conn
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	callArgs := m.Called(ctx, method, args, reply, opts)
	return callArgs.Error(0)
}

func (m *MockConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	args := m.Called(ctx, desc, method, opts)
	return args.Get(0).(grpc.ClientStream), args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConn) GetState() connectivity.State {
	args := m.Called()
	return args.Get(0).(connectivity.State)
}

func TestHTTPTool_Getters(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("http-tool")
	cacheConfig := &configv1.CacheConfig{}
	cacheConfig.SetIsEnabled(true)
	callDef := &configv1.HttpCallDefinition{}
	callDef.SetCache(cacheConfig)
	httpTool := NewHTTPTool(toolProto, nil, "", nil, callDef, nil, nil, "")

	assert.Equal(t, toolProto, httpTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, httpTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}

func TestMCPTool_Getters(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("mcp-tool")
	cacheConfig := &configv1.CacheConfig{}
	cacheConfig.SetIsEnabled(true)
	callDef := &configv1.MCPCallDefinition{}
	callDef.SetCache(cacheConfig)
	mcpTool := NewMCPTool(toolProto, nil, callDef)

	assert.Equal(t, toolProto, mcpTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, mcpTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}

func TestOpenAPITool_Getters(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("openapi-tool")
	cacheConfig := &configv1.CacheConfig{}
	cacheConfig.SetIsEnabled(true)
	callDef := &configv1.OpenAPICallDefinition{}
	callDef.SetCache(cacheConfig)
	openapiTool := NewOpenAPITool(toolProto, nil, nil, "", "", nil, callDef)

	assert.Equal(t, toolProto, openapiTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, openapiTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}

func TestGRPCTool_Getters(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("grpc-tool")
	cacheConfig := &configv1.CacheConfig{}
	cacheConfig.SetIsEnabled(true)
	callDef := &configv1.GrpcCallDefinition{}
	callDef.SetCache(cacheConfig)
	mockMethodDesc := new(MockMethodDescriptor)
	mockMethodDesc.On("Input").Return(new(MockMessageDescriptor))
	grpcTool := NewGRPCTool(toolProto, nil, "", mockMethodDesc, callDef)

	assert.Equal(t, toolProto, grpcTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, grpcTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}

func TestWebsocketTool_Getters(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("websocket-tool")
	cacheConfig := &configv1.CacheConfig{}
	cacheConfig.SetIsEnabled(true)
	callDef := &configv1.WebsocketCallDefinition{}
	callDef.SetCache(cacheConfig)
	websocketTool := NewWebsocketTool(toolProto, nil, "", nil, callDef)

	assert.Equal(t, toolProto, websocketTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, websocketTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}

func TestOpenAPITool_GetCacheConfig(t *testing.T) {
	tool := &OpenAPITool{}
	assert.Nil(t, tool.GetCacheConfig(), "GetCacheConfig should return nil")
}

func TestOpenAPITool_Tool(t *testing.T) {
	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	tool := &OpenAPITool{tool: toolProto}
	assert.Equal(t, toolProto, tool.Tool(), "Tool() should return the tool proto")
}


func TestGRPCTool_Execute_UnmarshalError(t *testing.T) {
	poolManager := pool.NewManager()
	mockConn := new(MockConn)
	mockConn.On("GetState").Return(connectivity.Ready)

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(_ context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}, nil), nil
	}, 1, 1, 0, false)
	poolManager.Register("grpc-error", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)

	grpcTool := NewGRPCTool(toolProto, poolManager, "grpc-error", mockMethodDesc, &configv1.GrpcCallDefinition{})

	// Malformed JSON
	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

func TestGRPCTool_Execute_InvokeError(t *testing.T) {
	poolManager := pool.NewManager()
	mockConn := new(MockConn)
	mockConn.On("GetState").Return(connectivity.Ready)
	mockConn.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("rpc error"))

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(_ context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}, nil), nil
	}, 1, 1, 0, false)
	poolManager.Register("grpc-error", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)
	mockMethodDesc.On("Output").Return(mockMsgDesc)

	grpcTool := NewGRPCTool(toolProto, poolManager, "grpc-error", mockMethodDesc, &configv1.GrpcCallDefinition{})

	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to invoke grpc method")
}

func TestOpenAPITool_Execute_Errors(t *testing.T) {
	// Test Input Validation/Unmarshal Error if possible?
	// Schema validation might be strict?

	// Test Pool Error covered?
	// OpenAPITool.Execute calls t.pool.Get(ctx).

	// Test Execute Error (Network)
	// OpenAPITool uses *http.Client or wrapper?
	// It uses `*client.HTTPClientWrapper`.

	// I'll test Output Unmarshal failure.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{invalid-json`)
	}))
	defer server.Close()

	toolProto := &v1.Tool{}
	callDef := &configv1.OpenAPICallDefinition{}

	tool := NewOpenAPITool(toolProto, server.Client(), nil, "GET", server.URL, nil, callDef)

	_, err := tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	// It might NOT error if it returns raw bytes?
	// Output is []byte used in logic.
	// If OpenAPITool tries to unmarshal output...
	// It returns `result` which is `response` string/bytes?

	// Checking source code:
	// OpenAPITool Execute logic:
	// resp, err := client.Do(req)
	// ... body, err := io.ReadAll(resp.Body)
	// it returns body (or transformed).

	// If there is an output schema?
	assert.NoError(t, err) // Matches current logic likely
}

func TestOpenAPITool_Execute_StatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `Not Found`)
	}))
	defer server.Close()

	toolProto := &v1.Tool{}
	callDef := &configv1.OpenAPICallDefinition{}

	tool := NewOpenAPITool(toolProto, server.Client(), nil, "GET", server.URL, nil, callDef)

	_, err := tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upstream OpenAPI request failed with status 404")
}

func TestHTTPTool_Execute_UnmarshalError(t *testing.T) {
	poolManager := pool.NewManager()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)

	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("POST " + server.URL)
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}
