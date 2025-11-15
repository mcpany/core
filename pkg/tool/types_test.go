/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
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

func (m *MockMCPClient) ExecuteTool(ctx context.Context, req *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

func TestContextWithTool(t *testing.T) {
	ctx := context.Background()
	mockTool := new(tool.MockTool)
	ctx = tool.NewContextWithTool(ctx, mockTool)
	retrievedTool, ok := tool.GetFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, mockTool, retrievedTool)

	_, ok = tool.GetFromContext(context.Background())
	assert.False(t, ok)
}

func TestGRPCTool_Execute_PoolError(t *testing.T) {
	poolManager := pool.NewManager()
	toolProto := &v1.Tool{}
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)
	grpcTool := tool.NewGRPCTool(toolProto, poolManager, "non-existent-service", mockMethodDesc, &configv1.GrpcCallDefinition{})
	_, err := grpcTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no grpc pool found for service")
}

func TestHTTPTool_Execute_PoolError(t *testing.T) {
	poolManager := pool.NewManager()
	toolProto := &v1.Tool{}
	httpTool := tool.NewHTTPTool(toolProto, poolManager, "non-existent-service", nil, &configv1.HttpCallDefinition{})
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no http pool found for service")
}

func TestHTTPTool_Execute_InvalidFQN(t *testing.T) {
	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HttpClientWrapper](func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("invalid")
	httpTool := tool.NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{})
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http tool definition")
}

func TestHTTPTool_Execute_BadURL(t *testing.T) {
	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HttpClientWrapper](func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("GET %")
	httpTool := tool.NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{})
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
}

func TestHTTPTool_Execute_InputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HttpClientWrapper](func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)

	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("POST " + server.URL)
	callDef := &configv1.HttpCallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetInputTransformer(inputTransformer)
	httpTool := tool.NewHTTPTool(toolProto, poolManager, "http-service", nil, callDef)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestHTTPTool_Execute_OutputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer server.Close()

	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HttpClientWrapper](func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)

	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("GET " + server.URL)
	callDef := &configv1.HttpCallDefinition{}
	outputTransformer := &configv1.OutputTransformer{}
	outputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetOutputTransformer(outputTransformer)
	httpTool := tool.NewHTTPTool(toolProto, poolManager, "http-service", nil, callDef)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
}

func TestMCPTool_Execute_InputTransformerError(t *testing.T) {
	toolProto := &v1.Tool{}
	callDef := &configv1.MCPCallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetInputTransformer(inputTransformer)
	mcpTool := tool.NewMCPTool(toolProto, nil, callDef)
	_, err := mcpTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestMCPTool_Execute_OutputTransformerError(t *testing.T) {
	mockClient := new(MockMCPClient)
	mockResult := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: `{"key":"value"}`}},
	}
	mockClient.On("ExecuteTool", mock.Anything, mock.Anything).Return(mockResult, nil)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	callDef := &configv1.MCPCallDefinition{}
	outputTransformer := &configv1.OutputTransformer{}
	outputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetOutputTransformer(outputTransformer)
	mcpTool := tool.NewMCPTool(toolProto, mockClient, callDef)
	_, err := mcpTool.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "test.test-tool", ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestOpenAPITool_Execute_InputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	toolProto := &v1.Tool{}
	callDef := &configv1.OpenAPICallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetInputTransformer(inputTransformer)
	openapiTool := tool.NewOpenAPITool(toolProto, server.Client(), nil, "POST", server.URL, nil, callDef)
	_, err := openapiTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{"key":"value"}`)})
	assert.Error(t, err)
}

func TestOpenAPITool_Execute_OutputTransformerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer server.Close()

	toolProto := &v1.Tool{}
	callDef := &configv1.OpenAPICallDefinition{}
	outputTransformer := &configv1.OutputTransformer{}
	outputTransformer.SetTemplate("{{.invalid}}")
	callDef.SetOutputTransformer(outputTransformer)
	openapiTool := tool.NewOpenAPITool(toolProto, server.Client(), nil, "GET", server.URL, nil, callDef)
	_, err := openapiTool.Execute(context.Background(), &tool.ExecutionRequest{})
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

func (m *MockFieldNumbers) Get(i int) protoreflect.FieldNumber {
	panic("should not be called")
}

func (m *MockFieldNumbers) Has(n protoreflect.FieldNumber) bool {
	return false
}

func TestGRPCTool_Execute_Success(t *testing.T) {
	poolManager := pool.NewManager()

	mockConn := new(MockConn)
	mockConn.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockConn.On("GetState").Return(connectivity.Ready)

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(ctx context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}), nil
	}, 1, 1, 0, false)
	poolManager.Register("grpc-service", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)

	mockMethodDesc.On("Input").Return(mockMsgDesc)
	mockMethodDesc.On("Output").Return(mockMsgDesc)

	grpcTool := tool.NewGRPCTool(toolProto, poolManager, "grpc-service", mockMethodDesc, &configv1.GrpcCallDefinition{})
	_, err := grpcTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
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
	httpTool := tool.NewHTTPTool(toolProto, nil, "", nil, callDef)

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
	mcpTool := tool.NewMCPTool(toolProto, nil, callDef)

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
	openapiTool := tool.NewOpenAPITool(toolProto, nil, nil, "", "", nil, callDef)

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
	grpcTool := tool.NewGRPCTool(toolProto, nil, "", mockMethodDesc, callDef)

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
	websocketTool := tool.NewWebsocketTool(toolProto, nil, "", nil, callDef)

	assert.Equal(t, toolProto, websocketTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, websocketTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}
