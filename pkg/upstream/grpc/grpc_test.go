/*
 * Copyright 2025 Author(s) of MCP-XY
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

package grpc

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream/grpc/protobufparser"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/examples/calculator/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// MockToolManager is a mock implementation of the ToolManagerInterface.
type MockToolManager struct {
	mu           sync.Mutex
	tools        map[string]tool.Tool
	serviceInfos map[string]*tool.ServiceInfo
	lastErr      error
}

func NewMockToolManager() *MockToolManager {
	return &MockToolManager{
		tools:        make(map[string]tool.Tool),
		serviceInfos: make(map[string]*tool.ServiceInfo),
	}
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastErr != nil {
		return m.lastErr
	}
	toolID, _ := util.GenerateToolID(t.Tool().GetServiceId(), t.Tool().GetName())
	m.tools[toolID] = t
	return nil
}

func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tools[name]
	return t, ok
}

func (m *MockToolManager) ListTools() []tool.Tool {
	m.mu.Lock()
	defer m.mu.Unlock()
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

func (m *MockToolManager) ClearToolsForService(serviceKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceKey {
			delete(m.tools, name)
		}
	}
}

func (m *MockToolManager) SetMCPServer(provider tool.MCPServerProvider) {}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serviceInfos[serviceID] = info
}

func (m *MockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.serviceInfos[serviceID]
	return info, ok
}

func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func TestNewGRPCUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &GRPCUpstream{}, upstream)
}

func TestGRPCUpstream_Register(t *testing.T) {
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	t.Run("nil service config", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		_, _, err := upstream.Register(context.Background(), nil, nil, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("") // empty name is invalid
		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress("localhost:50051")
		serviceConfig.SetGrpcService(grpcService)
		_, _, err := upstream.Register(context.Background(), serviceConfig, NewMockToolManager(), promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("nil grpc service config", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test")
		_, _, err := upstream.Register(context.Background(), serviceConfig, nil, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Equal(t, "grpc service config is nil", err.Error())
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test")
		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress("localhost:50051")
		serviceConfig.SetGrpcService(grpcService)
		authConfig := (&configv1.UpstreamAuthentication_builder{
			BearerToken: &configv1.UpstreamBearerTokenAuth{}, // Invalid config
		}).Build()
		serviceConfig.SetUpstreamAuthentication(authConfig)

		_, _, err := upstream.Register(context.Background(), serviceConfig, NewMockToolManager(), promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create upstream authenticator")
	})

	t.Run("reflection fails", func(t *testing.T) {
		// Start a simple HTTP server, not a gRPC server
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer httpServer.Close()

		// Extract host and port from the URL
		parsedURL, err := url.Parse(httpServer.URL)
		require.NoError(t, err)

		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(parsedURL.Host)
		grpcService.SetUseReflection(true)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("reflection-fail-service")
		serviceConfig.SetGrpcService(grpcService)

		_, _, err = upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to discover service by reflection")
	})
}

func TestGRPCUpstream_createAndRegisterGRPCTools(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := NewMockToolManager()

	t.Run("nil parsed data", func(t *testing.T) {
		tools, err := upstream.(*GRPCUpstream).createAndRegisterGRPCTools(context.Background(), "test-service", nil, tm, false, nil)
		require.NoError(t, err)
		assert.Nil(t, tools)
	})

	t.Run("service info not found", func(t *testing.T) {
		parsedData := &protobufparser.ParsedMcpAnnotations{}
		_, err := upstream.(*GRPCUpstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service info not found")
	})

	t.Run("bad file descriptor set", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		tm := NewMockToolManager()
		tm.AddServiceInfo("test-service", &tool.ServiceInfo{})

		parsedData := &protobufparser.ParsedMcpAnnotations{
			Tools: []protobufparser.McpTool{
				{Name: "test-tool"},
			},
		}
		// Create a malformed FileDescriptorSet with a missing dependency
		fds := &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:       proto.String("test.proto"),
					Dependency: []string{"nonexistent.proto"},
				},
			},
		}

		_, err := upstream.(*GRPCUpstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, false, fds)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create protodesc files")
	})
}

// Mock gRPC server for testing
type mockCalculatorServer struct {
	pb.UnimplementedCalculatorServiceServer
}

func (s *mockCalculatorServer) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	result := in.GetA() + in.GetB()
	return pb.AddResponse_builder{Result: &result}.Build(), nil
}

func (s *mockCalculatorServer) Subtract(ctx context.Context, in *pb.SubtractRequest) (*pb.SubtractResponse, error) {
	result := in.GetA() - in.GetB()
	return pb.SubtractResponse_builder{Result: &result}.Build(), nil
}

func startMockServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(s, &mockCalculatorServer{})
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("mock server stopped: %v", err)
		}
	}()
	return s, lis.Addr().String()
}

func TestGRPCUpstream_Register_WithMockServer(t *testing.T) {
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	t.Run("successful registration with reflection and cache hit", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(addr)
		grpcService.SetUseReflection(true)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("calculator-service")
		serviceConfig.SetGrpcService(grpcService)

		// First call - should populate the cache
		serviceKey, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.NotEmpty(t, discoveredTools)
		assert.Len(t, tm.ListTools(), 2)

		// Second call - should hit the cache
		tm2 := NewMockToolManager()
		serviceKey2, discoveredTools2, err2 := upstream.Register(context.Background(), serviceConfig, tm2, promptManager, resourceManager, false)
		require.NoError(t, err2)
		assert.NotEmpty(t, discoveredTools2)
		assert.Len(t, tm2.ListTools(), 2)
		assert.Equal(t, serviceKey, serviceKey2)
		// We can't directly verify the cache was hit without exporting the cache,
		// but a successful second call is a good indicator.
	})
}

func TestFindMethodDescriptor(t *testing.T) {
	server, addr := startMockServer(t)
	defer server.Stop()
	ctx := context.Background()
	fds, err := protobufparser.ParseProtoByReflection(ctx, addr)
	require.NoError(t, err)
	files, err := protodesc.NewFiles(fds)
	require.NoError(t, err)

	t.Run("invalid full method name", func(t *testing.T) {
		_, err := findMethodDescriptor(files, "invalidname")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid full method name")
	})

	t.Run("service not found", func(t *testing.T) {
		_, err := findMethodDescriptor(files, "nonexistent.Service/Method")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not find descriptor for service 'nonexistent.Service'")
	})

	t.Run("descriptor is not a service", func(t *testing.T) {
		// Use a message type instead of a service type
		_, err := findMethodDescriptor(files, "calculator.v1.AddRequest/Method")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not find descriptor for service 'calculator.v1.AddRequest'")
	})

	t.Run("method not found", func(t *testing.T) {
		_, err := findMethodDescriptor(files, "calculator.v1.CalculatorService/NonExistentMethod")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not find descriptor for service 'calculator.v1.CalculatorService'")
	})
}

func TestConvertMcpFieldsToInputSchemaProperties(t *testing.T) {
	fields := []*protobufparser.McpField{
		{Name: "field1", Description: "a string field", Type: "TYPE_STRING"},
		{Name: "field2", Description: "an int field", Type: "TYPE_INT32"},
		{Name: "field3", Description: "a bool field", Type: "TYPE_BOOL"},
		{Name: "field4", Description: "a float field", Type: "TYPE_FLOAT"},
		{Name: "field5", Description: "", Type: "TYPE_STRING"},
	}

	properties, err := convertMcpFieldsToInputSchemaProperties(fields)
	require.NoError(t, err)
	assert.Len(t, properties, 5)

	// Check field1 - should be a struct with type and description
	field1, ok := properties["field1"]
	require.True(t, ok)
	s1, ok := field1.GetKind().(*structpb.Value_StructValue)
	require.True(t, ok)
	assert.Equal(t, "string", s1.StructValue.Fields["type"].GetStringValue())
	assert.Equal(t, "a string field", s1.StructValue.Fields["description"].GetStringValue())

	// Check field2
	field2, ok := properties["field2"]
	require.True(t, ok)
	s2, ok := field2.GetKind().(*structpb.Value_StructValue)
	require.True(t, ok)
	assert.Equal(t, "integer", s2.StructValue.Fields["type"].GetStringValue())

	// Check field3
	field3, ok := properties["field3"]
	require.True(t, ok)
	s3, ok := field3.GetKind().(*structpb.Value_StructValue)
	require.True(t, ok)
	assert.Equal(t, "boolean", s3.StructValue.Fields["type"].GetStringValue())

	// Check field4
	field4, ok := properties["field4"]
	require.True(t, ok)
	s4, ok := field4.GetKind().(*structpb.Value_StructValue)
	require.True(t, ok)
	assert.Equal(t, "number", s4.StructValue.Fields["type"].GetStringValue())

	// Check field5
	field5, ok := properties["field5"]
	require.True(t, ok)
	s5, ok := field5.GetKind().(*structpb.Value_StructValue)
	require.True(t, ok)
	assert.Equal(t, "string", s5.StructValue.Fields["type"].GetStringValue())
	assert.Equal(t, "", s5.StructValue.Fields["description"].GetStringValue())
}
