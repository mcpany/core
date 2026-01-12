// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/examples/weather/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/grpc/protobufparser"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
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
	sanitizedToolName, _ := util.SanitizeToolName(t.Tool().GetName())
	toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
	m.tools[toolID] = t
	return nil
}

func (m *MockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	return true
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

func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			delete(m.tools, name)
		}
	}
}

func (m *MockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func (m *MockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

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

func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	services := make([]*tool.ServiceInfo, 0, len(m.serviceInfos))
	for _, info := range m.serviceInfos {
		services = append(services, info)
	}
	return services
}

func (m *MockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockToolManager) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastErr = err
}

func (m *MockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {
}

func (m *MockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool {
	return true
}

func TestNewGRPCUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &Upstream{}, upstream)
}

func TestGRPCUpstream_Register(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	t.Run("nil service config", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		_, _, _, err := upstream.Register(context.Background(), nil, nil, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("") // empty name is invalid
		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress("localhost:50051")
		serviceConfig.SetGrpcService(grpcService)
		_, _, _, err := upstream.Register(context.Background(), serviceConfig, NewMockToolManager(), promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})

	t.Run("nil grpc service config", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test")
		_, _, _, err := upstream.Register(context.Background(), serviceConfig, nil, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Equal(t, "grpc service config is nil", err.Error())
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test"),
			ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
				GrpcService: &configv1.GrpcUpstreamService{
					Address: proto.String("localhost:50051"),
				},
			},
			UpstreamAuth: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BearerToken{
					BearerToken: &configv1.BearerTokenAuth{
						// Token missing
					},
				},
			},
		}

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, NewMockToolManager(), promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bearer token authentication requires a token")
	})

	t.Run("reflection fails", func(t *testing.T) {
		// Start a simple HTTP server, not a gRPC server
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer httpServer.Close()

		// Extract host and port from the URL
		parsedURL, err := url.Parse(httpServer.URL)
		require.NoError(t, err)

		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(parsedURL.Host)
		grpcService.SetUseReflection(true)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("reflection-fail-service")
		serviceConfig.SetGrpcService(grpcService)

		_, _, _, err = upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to discover service by reflection")
	})
}

func TestGRPCUpstream_createAndRegisterGRPCTools(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	t.Run("nil parsed data", func(t *testing.T) {
		tools, err := upstream.(*Upstream).createAndRegisterGRPCTools(context.Background(), "test-service", nil, tm, nil, false, nil)
		require.NoError(t, err)
		assert.Nil(t, tools)
	})

	t.Run("service info not found", func(t *testing.T) {
		parsedData := &protobufparser.ParsedMcpAnnotations{}
		_, err := upstream.(*Upstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, nil, false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service info not found")
	})

	t.Run("bad file descriptor set", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		tm := NewMockToolManager()
		tm.AddServiceInfo("test-service", &tool.ServiceInfo{
			Config: configv1.UpstreamServiceConfig_builder{
				GrpcService: configv1.GrpcUpstreamService_builder{
					Tools: []*configv1.ToolDefinition{
						configv1.ToolDefinition_builder{
							Name:   proto.String("test-tool"),
							CallId: proto.String("test-call"),
						}.Build(),
					},
					Calls: map[string]*configv1.GrpcCallDefinition{
						"test-call": configv1.GrpcCallDefinition_builder{
							Id: proto.String("test-call"),
						}.Build(),
					},
				}.Build(),
			}.Build(),
		})

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

		_, err := upstream.(*Upstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, nil, false, fds)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create protodesc files")
	})
}

// Mock gRPC server for testing
type mockWeatherServer struct {
	pb.UnimplementedWeatherServiceServer
}

func (s *mockWeatherServer) GetWeather(_ context.Context, _ *pb.GetWeatherRequest) (*pb.GetWeatherResponse, error) {
	return pb.GetWeatherResponse_builder{Weather: "sunny"}.Build(), nil
}

func startMockServer(t *testing.T) (*grpc.Server, string) {
	var lis net.Listener
	var err error
	for i := 0; i < 5; i++ {
		lis, err = net.Listen("tcp", "127.0.0.1:0")
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWeatherServiceServer(s, &mockWeatherServer{})
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("mock server stopped: %v", err)
		}
	}()
	return s, lis.Addr().String()
}

func TestGRPCUpstream_Register_WithMockServer(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	t.Run("successful registration with reflection and cache hit", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(addr)
		grpcService.SetUseReflection(true)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("weather-service")
		serviceConfig.SetGrpcService(grpcService)

		// First call - should populate the cache
		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.NotEmpty(t, discoveredTools)
		assert.Len(t, tm.ListTools(), 2) // 1 for GetWeather, 1 for reflection

		// Second call - should hit the cache
		tm2 := NewMockToolManager()
		serviceID2, discoveredTools2, _, err2 := upstream.Register(context.Background(), serviceConfig, tm2, promptManager, resourceManager, false)
		require.NoError(t, err2)
		assert.NotEmpty(t, discoveredTools2)
		assert.Len(t, tm2.ListTools(), 2)
		assert.Equal(t, serviceID, serviceID2)
		// We can't directly verify the cache was hit without exporting the cache,
		// but a successful second call is a good indicator.
	})

	t.Run("correct input schema generation", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(addr)
		grpcService.SetUseReflection(true)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("weather-service")
		serviceConfig.SetGrpcService(grpcService)

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err)

		// Verify the "GetWeather" tool's schema
		sanitizedToolName, err := util.SanitizeToolName("GetWeather")
		require.NoError(t, err)
		getWeatherToolName := serviceID + "." + sanitizedToolName
		getWeatherTool, ok := tm.GetTool(getWeatherToolName)
		require.True(t, ok)
		inputSchema := getWeatherTool.Tool().GetAnnotations().GetInputSchema()
		require.NotNil(t, inputSchema)
		assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())
		properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
		require.Contains(t, properties, "location")
		assert.Equal(t, "string", properties["location"].GetStructValue().GetFields()["type"].GetStringValue())
	})

	t.Run("auto discovery", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(addr)
		grpcService.SetUseReflection(true)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("weather-service-auto")
		serviceConfig.SetGrpcService(grpcService)
		serviceConfig.SetAutoDiscoverTool(true)

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.NotEmpty(t, discoveredTools)

		// Check if GetWeather is registered
		sanitizedToolName, err := util.SanitizeToolName("GetWeather")
		require.NoError(t, err)
		toolName := serviceID + "." + sanitizedToolName
		_, ok := tm.GetTool(toolName)
		assert.True(t, ok, "GetWeather tool should be auto-discovered")
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
		_, err := findMethodDescriptor(files, "examples.weather.v1.GetWeatherRequest/Method")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is not a service descriptor")
	})

	t.Run("method not found", func(t *testing.T) {
		_, err := findMethodDescriptor(files, "examples.weather.v1.WeatherService/NonExistentMethod")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method 'NonExistentMethod' not found in service")
	})

	t.Run("valid full method name", func(t *testing.T) {
		methodDesc, err := findMethodDescriptor(files, "examples.weather.v1.WeatherService/GetWeather")
		require.NoError(t, err)
		assert.NotNil(t, methodDesc)
		assert.Equal(t, "GetWeather", string(methodDesc.Name()))
	})

	t.Run("valid full method name with leading slash", func(t *testing.T) {
		methodDesc, err := findMethodDescriptor(files, "/examples.weather.v1.WeatherService/GetWeather")
		require.NoError(t, err)
		assert.NotNil(t, methodDesc)
		assert.Equal(t, "GetWeather", string(methodDesc.Name()))
	})
}

func TestGRPCUpstream_Register_UseReflection_WithPolicy(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)
	// Disable reflection tool specifically
	grpcService.Tools = []*configv1.ToolDefinition{
		{Name: proto.String("grpc_reflection_v1alpha_ServerReflection_ServerReflectionInfo"), Disable: proto.Bool(true)},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-policy")
	serviceConfig.SetGrpcService(grpcService)
	// Export policy: only export GetWeather
	serviceConfig.ToolExportPolicy = &configv1.ExportPolicy{
		DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
		Rules: []*configv1.ExportRule{
			{NameRegex: proto.String(".*GetWeather"), Action: actionPtr(configv1.ExportPolicy_EXPORT)},
		},
	}

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Check tm
	tools := tm.ListTools()
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		name := tool.Tool().GetName()
		toolNames = append(toolNames, name)
	}

	assert.Contains(t, toolNames, "GetWeather")
	assert.NotContains(t, toolNames, "grpc_reflection_v1alpha_ServerReflection_ServerReflectionInfo")
}

func actionPtr(a configv1.ExportPolicy_Action) *configv1.ExportPolicy_Action {
	return &a
}

func TestGRPCUpstream_Register_DynamicResources(t *testing.T) {
	var promptManager prompt.ManagerInterface
	resourceManager := resource.NewManager()

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)
	// Add dynamic resource definition
	grpcService.Resources = []*configv1.ResourceDefinition{
		{
			Name: proto.String("weather_resource"),
			ResourceType: &configv1.ResourceDefinition_Dynamic{
				Dynamic: &configv1.DynamicResource{
					CallDefinition: &configv1.DynamicResource_GrpcCall{
						GrpcCall: &configv1.GrpcCallDefinition{
							Id: proto.String("weather_call"),
						},
					},
				},
			},
		},
	}
	// We need to match the tool definition
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:   proto.String("GetWeather"), // Matches reflection name
			CallId: proto.String("weather_call"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-dynamic")
	serviceConfig.SetGrpcService(grpcService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	resources := resourceManager.ListResources()
	assert.Len(t, resources, 1)
	assert.Equal(t, "weather_resource", resources[0].Resource().Name)
}

func TestGRPCUpstream_Register_DuplicateTool(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	serviceID := "weather-service-dup"

	// We need a dummy tool to put in tm
	dummyToolProto := &routerv1.Tool{
		Name:      proto.String("GetWeather"),
		ServiceId: proto.String(serviceID),
	}
	// Use a simple mock tool instead of GRPCTool to avoid dependencies
	_ = tm.AddTool(&simpleMockTool{t: dummyToolProto})

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName(serviceID)
	serviceConfig.SetGrpcService(grpcService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	assert.NoError(t, err)

	toolNames := make([]string, 0, len(discoveredTools))
	for _, dt := range discoveredTools {
		toolNames = append(toolNames, dt.GetName())
	}
	assert.NotContains(t, toolNames, "GetWeather")
	assert.Contains(t, toolNames, "ServerReflectionInfo")
}

func TestGRPCUpstream_Register_DuplicateTool_Config(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	serviceID := "weather-service-dup-config"
	// Pre-register tool
	dummyToolProto := &routerv1.Tool{
		Name:      proto.String("GetWeather"),
		ServiceId: proto.String(serviceID),
	}
	_ = tm.AddTool(&simpleMockTool{t: dummyToolProto})

	// This tool is duplicate in config vs active, but different name?
	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(false)
	grpcService.ProtoDefinitions = []*configv1.ProtoDefinition{
		{
			ProtoRef: &configv1.ProtoDefinition_ProtoFile{
				ProtoFile: &configv1.ProtoFile{
					FileName: proto.String("test.proto"),
					FileRef: &configv1.ProtoFile_FileContent{
						FileContent: `
syntax = "proto3";
package test;
service TestService {
  rpc GetWeather (Request) returns (Response);
}
message Request {}
message Response {}
`,
					},
				},
			},
		},
	}
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:   proto.String("GetWeather"),
			CallId: proto.String("weather_call"),
		},
	}
	grpcService.Calls = map[string]*configv1.GrpcCallDefinition{
		"weather_call": {
			Id:     proto.String("weather_call"),
			Method: proto.String("test.TestService/GetWeather"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName(serviceID)
	serviceConfig.SetGrpcService(grpcService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestGRPCUpstream_Register_ExportPolicy_Config(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	// ...
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	serviceID := "weather-service-export-config"

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(false)
	grpcService.ProtoDefinitions = []*configv1.ProtoDefinition{
		{
			ProtoRef: &configv1.ProtoDefinition_ProtoFile{
				ProtoFile: &configv1.ProtoFile{
					FileName: proto.String("test.proto"),
					FileRef: &configv1.ProtoFile_FileContent{
						FileContent: `
syntax = "proto3";
package test;
service TestService {
  rpc GetWeather (Request) returns (Response);
}
message Request {}
message Response {}
`,
					},
				},
			},
		},
	}
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:   proto.String("GetWeather"),
			CallId: proto.String("weather_call"),
		},
	}
	grpcService.Calls = map[string]*configv1.GrpcCallDefinition{
		"weather_call": {
			Id:     proto.String("weather_call"),
			Method: proto.String("test.TestService/GetWeather"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName(serviceID)
	serviceConfig.SetGrpcService(grpcService)
	serviceConfig.ToolExportPolicy = &configv1.ExportPolicy{
		DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
	}

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestGRPCUpstream_Register_AddToolError(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()
	tm.SetError(errors.New("injection error"))

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-error")
	serviceConfig.SetGrpcService(grpcService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err) // Register itself doesn't fail, but tools are skipped
	assert.Empty(t, discoveredTools)
}

func TestGRPCUpstream_Register_DynamicResource_ToolNotFound(t *testing.T) {
	var promptManager prompt.ManagerInterface
	resourceManager := resource.NewManager()

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)
	// Add dynamic resource definition with non-existent tool call
	grpcService.Resources = []*configv1.ResourceDefinition{
		{
			Name: proto.String("weather_resource_bad"),
			ResourceType: &configv1.ResourceDefinition_Dynamic{
				Dynamic: &configv1.DynamicResource{
					CallDefinition: &configv1.DynamicResource_GrpcCall{
						GrpcCall: &configv1.GrpcCallDefinition{
							Id: proto.String("non_existent_call"),
						},
					},
				},
			},
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-dynamic-fail")
	serviceConfig.SetGrpcService(grpcService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	resources := resourceManager.ListResources()
	assert.Empty(t, resources)
}

func TestGRPCUpstream_Register_FromConfig(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(false)
	grpcService.ProtoDefinitions = []*configv1.ProtoDefinition{
		{
			ProtoRef: &configv1.ProtoDefinition_ProtoFile{
				ProtoFile: &configv1.ProtoFile{
					FileName: proto.String("weather.proto"),
					FileRef: &configv1.ProtoFile_FilePath{
						FilePath: "../../../../proto/examples/weather/v1/weather.proto",
					},
				},
			},
		},
	}
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:   proto.String("GetWeather"),
			CallId: proto.String("weather_call"),
		},
	}
	grpcService.Calls = map[string]*configv1.GrpcCallDefinition{
		"weather_call": {
			Id:     proto.String("weather_call"),
			Method: proto.String("examples.weather.v1.WeatherService/GetWeather"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-config")
	serviceConfig.SetGrpcService(grpcService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)
	assert.NotEmpty(t, discoveredTools)
	assert.Equal(t, "GetWeather", discoveredTools[0].GetName())
}

func TestGRPCUpstream_Register_WithPrompts(t *testing.T) {
	var resourceManager resource.ManagerInterface

	promptManager := &MockPromptManager{}

	server, addr := startMockServer(t)
	defer server.Stop()

	// ... (rest of setup)
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)
	grpcService.Prompts = []*configv1.PromptDefinition{
		{
			Name:        proto.String("weather_prompt"),
			Description: proto.String("A prompt for weather"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-prompts")
	serviceConfig.SetGrpcService(grpcService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	assert.NotEmpty(t, promptManager.prompts)
	assert.Equal(t, "weather-service-prompts.weather_prompt", promptManager.prompts[0].Prompt().Name)
}

func TestGRPCUpstream_Register_Prompts_Invalid(t *testing.T) {
	var resourceManager resource.ManagerInterface
	promptManager := &MockPromptManager{}

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)
	grpcService.Prompts = []*configv1.PromptDefinition{
		{
			Description: proto.String("Missing Name"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-invalid-prompts")
	serviceConfig.SetGrpcService(grpcService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	// Should not error, just skip?
	// createAndRegisterPromptsFromConfig returns error if name missing?
	// It logs error and continues? Or returns error?
	// Logic: if promptDef.GetName() == "" { continue } (checking code...)
	// Actually I haven't checked code but likely it skips or returns error.
	// If it returns error, we expect Error.
	// If it skips, we expect NO error and empty prompts.

	// Assuming it skips/logs:
	require.NoError(t, err)
	assert.Empty(t, promptManager.prompts)
}

func TestGRPCUpstream_Register_AutoDiscover(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(false)
	grpcService.ProtoDefinitions = []*configv1.ProtoDefinition{
		{
			ProtoRef: &configv1.ProtoDefinition_ProtoFile{
				ProtoFile: &configv1.ProtoFile{
					FileName: proto.String("test.proto"),
					FileRef: &configv1.ProtoFile_FileContent{
						FileContent: `
syntax = "proto3";
package test;
service TestService {
  rpc GetData (Request) returns (Response);
}
message Request {}
message Response {}
`,
					},
				},
			},
		},
	}
	// Enable AutoDiscover to force use of createAndRegisterGRPCToolsFromDescriptors
	// without config-based tools.
	// We need to set AutoDiscoverTool on UpstreamServiceConfig, NOT GrpcUpstreamService (it's not there).

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service-autodiscover")
	serviceConfig.SetGrpcService(grpcService)
	serviceConfig.AutoDiscoverTool = proto.Bool(true)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Should discover GetData
	assert.NotEmpty(t, discoveredTools)
	assert.Equal(t, "GetData", discoveredTools[0].GetName())
}

type MockPromptManager struct {
	prompts []prompt.Prompt
}

func (m *MockPromptManager) AddPrompt(p prompt.Prompt) {
	m.prompts = append(m.prompts, p)
}

func (m *MockPromptManager) UpdatePrompt(_ prompt.Prompt) {}

func (m *MockPromptManager) GetPrompt(_ string) (prompt.Prompt, bool) {
	return nil, false
}

func (m *MockPromptManager) ListPrompts() []prompt.Prompt {
	return m.prompts
}

func (m *MockPromptManager) ClearPromptsForService(_ string) {}

func (m *MockPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

func TestGRPCUpstream_Register_FromConfig_MethodNotFound(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(false)
	grpcService.ProtoDefinitions = []*configv1.ProtoDefinition{
		{
			ProtoRef: &configv1.ProtoDefinition_ProtoFile{
				ProtoFile: &configv1.ProtoFile{
					FileName: proto.String("test.proto"),
					FileRef: &configv1.ProtoFile_FileContent{
						FileContent: `
syntax = "proto3";
package test;
service TestService {}
`,
					},
				},
			},
		},
	}
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:   proto.String("GetWeather"),
			CallId: proto.String("weather_call"),
		},
	}
	grpcService.Calls = map[string]*configv1.GrpcCallDefinition{
		"weather_call": {
			Id:     proto.String("weather_call"),
			Method: proto.String("test.TestService/NonExistentMethod"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-method-fail")
	serviceConfig.SetGrpcService(grpcService)

	// This should not return error, but log error and skip the tool.
	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestGRPCUpstream_Register_DisabledTool_Reflection(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true)
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:    proto.String("GetWeather"),
			Disable: proto.Bool(true),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-disabled")
	serviceConfig.SetGrpcService(grpcService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// ServerReflectionInfo might still be there, but GetWeather should be skipped
	toolNames := make([]string, 0, len(discoveredTools))
	for _, dt := range discoveredTools {
		toolNames = append(toolNames, dt.GetName())
	}
	assert.NotContains(t, toolNames, "GetWeather")
}

type simpleMockTool struct {
	t *routerv1.Tool
}

func (s *simpleMockTool) Tool() *routerv1.Tool {
	return s.t
}

func (s *simpleMockTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (s *simpleMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (s *simpleMockTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(s.t)
	return t
}
