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

package mcpserver

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/worker"
	v1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/examples/calculator/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

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

func TestRegistrationServer_RegisterService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup bus and worker
	busProvider := bus.NewBusProvider()

	// Setup components
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)
	registrationWorker.Start(ctx)

	// Setup server
	registrationServer, err := NewRegistrationServer(busProvider)
	require.NoError(t, err)

	t.Run("successful registration", func(t *testing.T) {
		serviceName := "testservice"
		config := &configv1.UpstreamServiceConfig{}
		config.SetName(serviceName)
		config.SetHttpService(&configv1.HttpUpstreamService{})

		req := &v1.RegisterServiceRequest{}
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
		req := &v1.RegisterServiceRequest{}
		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("missing config name", func(t *testing.T) {
		config := &configv1.UpstreamServiceConfig{}
		req := &v1.RegisterServiceRequest{}
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

		serviceName := "calculator-service"
		useReflection := true
		grpcService := configv1.GrpcUpstreamService_builder{
			Address:       &addr,
			UseReflection: &useReflection,
		}.Build()

		config := configv1.UpstreamServiceConfig_builder{
			Name:        &serviceName,
			GrpcService: grpcService,
		}.Build()

		req := &v1.RegisterServiceRequest{}
		req.SetConfig(config)

		resp, err := registrationServer.RegisterService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		serviceID := resp.GetServiceKey()
		tools := toolManager.ListTools()
		// There will be other tools from other tests, so we need to find our tools
		var addTool tool.Tool
		for _, t := range tools {
			if t.Tool().GetServiceId() == serviceID && t.Tool().GetName() == "CalculatorAdd" {
				addTool = t
				break
			}
		}
		require.NotNil(t, addTool)

		inputSchema := addTool.Tool().GetAnnotations().GetInputSchema()
		require.NotNil(t, inputSchema)
		assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())
		properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
		require.Contains(t, properties, "a")
		require.Contains(t, properties, "b")
		assert.Equal(t, "integer", properties["a"].GetStructValue().GetFields()["type"].GetStringValue())
		assert.Equal(t, "integer", properties["b"].GetStructValue().GetFields()["type"].GetStringValue())
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
			OpenapiSpec: &spec,
		}.Build()

		config := configv1.UpstreamServiceConfig_builder{
			Name:           &serviceName,
			OpenapiService: openapiService,
		}.Build()

		req := &v1.RegisterServiceRequest{}
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
			Schema: configv1.ToolSchema_builder{
				Name: proto.String("test-tool"),
			}.Build(),
			Parameters: []*configv1.WebsocketParameterMapping{param1},
		}.Build()

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://localhost:8080/test"),
			Calls:   []*configv1.WebsocketCallDefinition{callDef},
		}.Build()

		config := configv1.UpstreamServiceConfig_builder{
			Name:             &serviceName,
			WebsocketService: websocketService,
		}.Build()

		req := &v1.RegisterServiceRequest{}
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
}

func TestRegistrationServer_Unimplemented(t *testing.T) {
	ctx := context.Background()
	busProvider := bus.NewBusProvider()
	registrationServer, err := NewRegistrationServer(busProvider)
	require.NoError(t, err)

	t.Run("UnregisterService", func(t *testing.T) {
		_, err := registrationServer.UnregisterService(ctx, &v1.UnregisterServiceRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})

	t.Run("InitiateOAuth2Flow", func(t *testing.T) {
		_, err := registrationServer.InitiateOAuth2Flow(ctx, &v1.InitiateOAuth2FlowRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})

	t.Run("RegisterTools", func(t *testing.T) {
		_, err := registrationServer.RegisterTools(ctx, &v1.RegisterToolsRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})

	t.Run("GetServiceStatus", func(t *testing.T) {
		_, err := registrationServer.GetServiceStatus(ctx, &v1.GetServiceStatusRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})
}
