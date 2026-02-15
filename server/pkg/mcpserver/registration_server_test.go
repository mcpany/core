package mcpserver

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	v1 "github.com/mcpany/core/proto/api/v1"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/examples/weather/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Mock gRPC server for testing
type mockWeatherServer struct {
	pb.UnimplementedWeatherServiceServer
}

func (s *mockWeatherServer) GetWeather(_ context.Context, _ *pb.GetWeatherRequest) (*pb.GetWeatherResponse, error) {
	return pb.GetWeatherResponse_builder{Weather: "sunny"}.Build(), nil
}

func startMockServer(t *testing.T) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", ":0") //nolint:gosec // test
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

func TestRegistrationServer_RegisterService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup bus and worker
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Setup components
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)
	registrationWorker.Start(ctx)

	// Setup server
	registrationServer, err := NewRegistrationServer(busProvider, authManager)
	require.NoError(t, err)

	t.Run("successful registration", func(t *testing.T) {
		serviceName := "testservice"
		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName(serviceName)
		httpService := configv1.HttpUpstreamService_builder{}.Build()
		httpService.SetAddress("http://127.0.0.1:8080")
		config.SetHttpService(httpService)

		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(config)

		resp, err := registrationServer.RegisterService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Contains(t, resp.GetMessage(), "registered successfully")

		// Verify that the service info was added to the tool manager
		serviceID := resp.GetServiceKey()
		require.NotEmpty(t, serviceID)
		serviceInfo, ok := toolManager.GetServiceInfo(serviceID)
		require.True(t, ok)
		require.NotNil(t, serviceInfo)
		assert.Equal(t, "testservice", serviceInfo.Name)
	})

	t.Run("missing config", func(t *testing.T) {
		req := v1.RegisterServiceRequest_builder{}.Build()
		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("missing config name", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{}.Build()
		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(config)
		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("grpc service with input schema", func(t *testing.T) {
		server, addr := startMockServer(t)
		defer server.Stop()

		serviceName := "weather-service"
		useReflection := true
		grpcService := configv1.GrpcUpstreamService_builder{
			Address:       &addr,
			UseReflection: &useReflection,
		}.Build()

		config := configv1.UpstreamServiceConfig_builder{
			Name:        &serviceName,
			GrpcService: grpcService,
		}.Build()

		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(config)

		resp, err := registrationServer.RegisterService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		serviceID := resp.GetServiceKey()
		tools := toolManager.ListTools()
		// There will be other tools from other tests, so we need to find our tools
		var getWeatherTool tool.Tool
		for _, t := range tools {
			if t.Tool().GetServiceId() == serviceID && t.Tool().GetName() == "GetWeather" {
				getWeatherTool = t
				break
			}
		}
		require.NotNil(t, getWeatherTool)

		inputSchema := getWeatherTool.Tool().GetAnnotations().GetInputSchema()
		require.NotNil(t, inputSchema)
		assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())
		properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
		require.Contains(t, properties, "location")
		assert.Equal(t, "string", properties["location"].GetStructValue().GetFields()["type"].GetStringValue())
	})

	t.Run("openapi service with input schema", func(t *testing.T) {
		spec := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
`
		serviceName := "openapi-service"
		openapiService := configv1.OpenapiUpstreamService_builder{
			SpecContent: &spec,
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("getUser"),
					CallId: proto.String("getUser-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.OpenAPICallDefinition{
				"getUser-call": configv1.OpenAPICallDefinition_builder{
					Id: proto.String("getUser-call"),
				}.Build(),
			},
		}.Build()

		config := configv1.UpstreamServiceConfig_builder{
			Name:           &serviceName,
			OpenapiService: openapiService,
		}.Build()

		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(config)

		resp, err := registrationServer.RegisterService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		serviceID := resp.GetServiceKey()
		tools := toolManager.ListTools()
		var openapiTool tool.Tool
		for _, t := range tools {
			if t.Tool().GetServiceId() == serviceID && t.Tool().GetName() == "getUser" {
				openapiTool = t
				break
			}
		}
		require.NotNil(t, openapiTool)

		inputSchema := openapiTool.Tool().GetAnnotations().GetInputSchema()
		require.NotNil(t, inputSchema)
		assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())
		properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
		require.Contains(t, properties, "userId")
		assert.Equal(t, "string", properties["userId"].GetStructValue().GetFields()["type"].GetStringValue())
	})

	t.Run("websocket service with input schema", func(t *testing.T) {
		serviceName := "websocket-service"
		param1 := configv1.WebsocketParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("param1"),
			}.Build(),
		}.Build()
		callDef := configv1.WebsocketCallDefinition_builder{
			Parameters: []*configv1.WebsocketParameterMapping{param1},
		}.Build()

		callDef.SetId("test-call")
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["test-call"] = callDef
		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://127.0.0.1:8080/test"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("test-tool"),
					CallId: proto.String("test-call"),
				}.Build(),
			},
			Calls: calls,
		}.Build()

		config := configv1.UpstreamServiceConfig_builder{
			Name:             &serviceName,
			WebsocketService: websocketService,
		}.Build()

		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(config)

		resp, err := registrationServer.RegisterService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		serviceID := resp.GetServiceKey()
		tools := toolManager.ListTools()
		var wsTool tool.Tool
		for _, t := range tools {
			if t.Tool().GetServiceId() == serviceID && t.Tool().GetName() == "test-tool" {
				wsTool = t
				break
			}
		}
		require.NotNil(t, wsTool)

		inputSchema := wsTool.Tool().GetAnnotations().GetInputSchema()
		require.NotNil(t, inputSchema)
		assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())
		properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
		require.Contains(t, properties, "param1")
	})

	t.Run("registration failure", func(t *testing.T) {
		serviceName := "failing-service"
		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName(serviceName)
		// This address is invalid and will cause the registration to fail
		httpService := configv1.HttpUpstreamService_builder{}.Build()
		httpService.SetAddress("http://invalid-url:port")
		config.SetHttpService(httpService)

		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(config)

		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "invalid config")
	})
}

func TestListServices(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup bus and worker
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Setup components
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)
	registrationWorker.Start(ctx)

	// Setup server
	registrationServer, err := NewRegistrationServer(busProvider, authManager)
	require.NoError(t, err)

	// Register a service to be listed
	serviceName := "test-service-for-listing"
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	config.SetName(serviceName)
	httpService := configv1.HttpUpstreamService_builder{}.Build()
	httpService.SetAddress("http://127.0.0.1:8080")
	config.SetHttpService(httpService)
	req := v1.RegisterServiceRequest_builder{}.Build()
	req.SetConfig(config)
	_, err = registrationServer.RegisterService(ctx, req)
	require.NoError(t, err)

	// Now, list the services
	listReq := v1.ListServicesRequest_builder{}.Build()
	resp, err := registrationServer.ListServices(ctx, listReq)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Check that the registered service is in the list
	found := false
	for _, serviceInfo := range resp.GetServices() {
		if serviceInfo.GetName() == serviceName {
			found = true
			break
		}
	}
	assert.True(t, found, "the newly registered service should be in the list")

	t.Run("list services with bus error", func(t *testing.T) {
		// To simulate a bus error, we'll use a new bus provider that isn't properly configured
		// for the worker to publish results to.
		errorBusProvider, err := bus.NewProvider(bus_pb.MessageBus_builder{}.Build())
		require.NoError(t, err)

		errorRegistrationServer, err := NewRegistrationServer(errorBusProvider, authManager)
		require.NoError(t, err)

		shortCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		_, err = errorRegistrationServer.ListServices(shortCtx, v1.ListServicesRequest_builder{}.Build())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.DeadlineExceeded, st.Code())
	})
}

func TestRegistrationServer_Unimplemented(t *testing.T) {
	ctx := context.Background()
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	registrationServer, err := NewRegistrationServer(busProvider, nil)
	require.NoError(t, err)

	t.Run("UnregisterService", func(t *testing.T) {
		_, err := registrationServer.UnregisterService(ctx, v1.UnregisterServiceRequest_builder{}.Build())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})



	t.Run("RegisterTools", func(t *testing.T) {
		_, err := registrationServer.RegisterTools(ctx, v1.RegisterToolsRequest_builder{}.Build())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})

	t.Run("GetServiceStatus", func(t *testing.T) {
		_, err := registrationServer.GetServiceStatus(ctx, v1.GetServiceStatusRequest_builder{}.Build())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})

	t.Run("mustEmbedUnimplementedRegistrationServiceServer", func(t *testing.T) {
		s := &RegistrationServer{}
		assert.NotPanics(t, s.mustEmbedUnimplementedRegistrationServiceServer)
	})
}

func TestNewRegistrationServer_NilBus(t *testing.T) {
	_, err := NewRegistrationServer(nil, nil)
	assert.Error(t, err)
}

func TestRegistrationServer_Timeouts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Setup bus without a worker to simulate a timeout
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	registrationServer, err := NewRegistrationServer(busProvider, nil)
	require.NoError(t, err)

	t.Run("RegisterService timeout", func(t *testing.T) {
		httpSvc := configv1.HttpUpstreamService_builder{}.Build()
		httpSvc.SetAddress("http://127.0.0.1:8080")
		cfg := configv1.UpstreamServiceConfig_builder{}.Build()
		cfg.SetName("timeout-test")
		cfg.SetHttpService(httpSvc)

		req := v1.RegisterServiceRequest_builder{}.Build()
		req.SetConfig(cfg)
		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.DeadlineExceeded, st.Code())
	})

	t.Run("ListServices timeout", func(t *testing.T) {
		_, err := registrationServer.ListServices(ctx, v1.ListServicesRequest_builder{}.Build())
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.DeadlineExceeded, st.Code())
	})
}

func TestGetService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup bus and worker
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Setup components
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)
	registrationWorker.Start(ctx)

	// Setup server
	registrationServer, err := NewRegistrationServer(busProvider, authManager)
	require.NoError(t, err)

	// Register a service to be retrieved
	serviceName := "test-service-for-get"
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	config.SetName(serviceName)
	httpService := configv1.HttpUpstreamService_builder{}.Build()
	httpService.SetAddress("http://127.0.0.1:8080")
	config.SetHttpService(httpService)
	req := v1.RegisterServiceRequest_builder{}.Build()
	req.SetConfig(config)
	_, err = registrationServer.RegisterService(ctx, req)
	require.NoError(t, err)

	t.Run("get existing service", func(t *testing.T) {
		req := v1.GetServiceRequest_builder{}.Build()
		req.SetServiceName(serviceName)
		resp, err := registrationServer.GetService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.GetService())
		assert.Equal(t, serviceName, resp.GetService().GetName())
	})

	t.Run("get non-existent service", func(t *testing.T) {
		req := v1.GetServiceRequest_builder{}.Build()
		req.SetServiceName("non-existent")
		_, err := registrationServer.GetService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("missing service name", func(t *testing.T) {
		req := v1.GetServiceRequest_builder{}.Build()
		_, err := registrationServer.GetService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("timeout", func(t *testing.T) {
		// Create a server connected to a bus with no workers to force timeout
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		emptyBusProvider, err := bus.NewProvider(messageBus)
		require.NoError(t, err)

		timeoutServer, err := NewRegistrationServer(emptyBusProvider, authManager)
		require.NoError(t, err)

		// Use a short context to simulate timeout
		shortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		req := v1.GetServiceRequest_builder{}.Build()
		req.SetServiceName(serviceName)
		_, err = timeoutServer.GetService(shortCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.DeadlineExceeded, st.Code())
	})
}
