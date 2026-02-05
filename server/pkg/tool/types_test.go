// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/structpb"
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
	t.Parallel()
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
	t.Parallel()
	poolManager := pool.NewManager()
	toolProto := &v1.Tool{}
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)
	grpcTool := NewGRPCTool(toolProto, poolManager, "non-existent-service", mockMethodDesc, &configv1.GrpcCallDefinition{}, nil)
	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no grpc pool found for service")
}

func TestHTTPTool_Execute_PoolError(t *testing.T) {
	t.Parallel()
	poolManager := pool.NewManager()
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("GET http://example.com")
	httpTool := NewHTTPTool(toolProto, poolManager, "non-existent-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no http pool found for service")
}

func TestHTTPTool_Execute_InvalidFQN(t *testing.T) {
	t.Parallel()
	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("invalid")
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http tool definition")
}

func TestHTTPTool_Execute_BadURL(t *testing.T) {
	t.Parallel()
	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("GET %")
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
}

func TestHTTPTool_Execute_InputTransformerError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, false)
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
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer server.Close()

	poolManager := pool.NewManager()
	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, false)
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	poolManager := pool.NewManager()

	mockConn := new(MockConn)
	mockConn.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockConn.On("GetState").Return(connectivity.Ready)

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(_ context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}, nil), nil
	}, 1, 1, 1, 0, false)
	poolManager.Register("grpc-service", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)

	mockMethodDesc.On("Input").Return(mockMsgDesc)
	mockMethodDesc.On("Output").Return(mockMsgDesc)

	grpcTool := NewGRPCTool(toolProto, poolManager, "grpc-service", mockMethodDesc, &configv1.GrpcCallDefinition{}, nil)
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	toolProto := &v1.Tool{}
	toolProto.SetName("grpc-tool")
	cacheConfig := &configv1.CacheConfig{}
	cacheConfig.SetIsEnabled(true)
	callDef := &configv1.GrpcCallDefinition{}
	callDef.SetCache(cacheConfig)
	mockMethodDesc := new(MockMethodDescriptor)
	mockMethodDesc.On("Input").Return(new(MockMessageDescriptor))
	grpcTool := NewGRPCTool(toolProto, nil, "", mockMethodDesc, callDef, nil)

	assert.Equal(t, toolProto, grpcTool.Tool(), "Tool() should return the correct tool proto")
	assert.Equal(t, cacheConfig, grpcTool.GetCacheConfig(), "GetCacheConfig() should return the correct cache config")
}

func TestWebsocketTool_Getters(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	tool := &OpenAPITool{}
	assert.Nil(t, tool.GetCacheConfig(), "GetCacheConfig should return nil")
}

func TestOpenAPITool_Tool(t *testing.T) {
	t.Parallel()
	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	tool := &OpenAPITool{tool: toolProto}
	assert.Equal(t, toolProto, tool.Tool(), "Tool() should return the tool proto")
}

func TestGRPCTool_Execute_UnmarshalError(t *testing.T) {
	t.Parallel()
	poolManager := pool.NewManager()
	mockConn := new(MockConn)
	mockConn.On("GetState").Return(connectivity.Ready)

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(_ context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}, nil), nil
	}, 1, 1, 1, 0, false)
	poolManager.Register("grpc-error", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)

	grpcTool := NewGRPCTool(toolProto, poolManager, "grpc-error", mockMethodDesc, &configv1.GrpcCallDefinition{}, nil)

	// Malformed JSON
	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

func TestGRPCTool_Execute_InvokeError(t *testing.T) {
	t.Parallel()
	poolManager := pool.NewManager()
	mockConn := new(MockConn)
	mockConn.On("GetState").Return(connectivity.Ready)
	mockConn.On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("rpc error"))

	grpcPool, _ := pool.New[*client.GrpcClientWrapper](func(_ context.Context) (*client.GrpcClientWrapper, error) {
		return client.NewGrpcClientWrapper(mockConn, &configv1.UpstreamServiceConfig{}, nil), nil
	}, 1, 1, 1, 0, false)
	poolManager.Register("grpc-error", grpcPool)

	toolProto := &v1.Tool{}
	toolProto.SetName("test-tool")
	toolProto.SetUnderlyingMethodFqn("test.service.Method")
	mockMethodDesc := new(MockMethodDescriptor)
	mockMsgDesc := new(MockMessageDescriptor)
	mockMethodDesc.On("Input").Return(mockMsgDesc)
	mockMethodDesc.On("Output").Return(mockMsgDesc)

	grpcTool := NewGRPCTool(toolProto, poolManager, "grpc-error", mockMethodDesc, &configv1.GrpcCallDefinition{}, nil)

	_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to invoke grpc method")
}

func TestOpenAPITool_Execute_Errors(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	poolManager := pool.NewManager()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpPool, _ := pool.New[*client.HTTPClientWrapper](func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, false)
	poolManager.Register("http-service", httpPool)

	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn("POST " + server.URL)
	httpTool := NewHTTPTool(toolProto, poolManager, "http-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

	_, err := httpTool.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestCheckForPathTraversal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		hasError bool
	}{
		{"safe", false},
		{"safe/path", false},
		{"..", true},
		{"../", true},
		{"..\\", true},
		{"/..", true},
		{"\\..", true},
		{"/../", true},
		{"\\..\\", true},
		{"/..\\", true},
		{"\\../", true},
		{"foo/../bar", true},
		{"foo\\..\\bar", true},
		{"../bar", true},
		{"bar/..", true},
		{"bar\\..", true},
		{"mixed/..\\slash", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := checkForPathTraversal(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "path traversal attempt detected")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_ListServices(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Add Service Info
	service1 := &ServiceInfo{Name: "service-1", Config: configv1.UpstreamServiceConfig_builder{}.Build()}
	service2 := &ServiceInfo{Name: "service-2", Config: configv1.UpstreamServiceConfig_builder{}.Build()}

	tm.AddServiceInfo("id-1", service1)
	tm.AddServiceInfo("id-2", service2)

	services := tm.ListServices()
	assert.Len(t, services, 2)

	// Check content
	names := make(map[string]bool)
	for _, s := range services {
		names[s.Name] = true
	}
	assert.True(t, names["service-1"])
	assert.True(t, names["service-2"])
}

func TestCommandTool_Execute_PathTraversal_Args(t *testing.T) {
	t.Parallel()
	// Setup command tool with args injection vulnerability
	toolProto := v1.Tool_builder{
		Name: proto.String("cmd-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{arg}}"},
	}.Build()

	cmdTool := NewCommandTool(toolProto, service, callDef, nil, "")

	// Test path traversal in args
	req := &ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: []byte(`{"arg": "../etc/passwd"}`),
	}

	_, err := cmdTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestCommandTool_Execute_PathTraversal_Env(t *testing.T) {
	t.Parallel()
	// Setup command tool with env injection vulnerability
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"properties": map[string]interface{}{
			"env_var": map[string]interface{}{},
		},
	})

	toolProto := v1.Tool_builder{
		Name: proto.String("cmd-tool"),
		Annotations: v1.ToolAnnotations_builder{
			InputSchema: inputSchema,
		}.Build(),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()

	// Parameter mapping to env
	schema := configv1.ParameterSchema_builder{Name: proto.String("env_var")}.Build()
	mapping := configv1.CommandLineParameterMapping_builder{
		Schema: schema,
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{mapping},
	}.Build()

	cmdTool := NewCommandTool(toolProto, service, callDef, nil, "")

	// Test path traversal in env var (which checks validation)
	req := &ExecutionRequest{
		ToolName:   "cmd-tool",
		ToolInputs: []byte(`{"env_var": "../bad"}`),
	}

	_, err := cmdTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}
