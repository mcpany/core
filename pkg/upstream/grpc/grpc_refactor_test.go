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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/proto/examples/calculator/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
)

type mockGRPCServer struct {
	v1.UnimplementedCalculatorServiceServer
}

func (s *mockGRPCServer) Add(ctx context.Context, in *v1.AddRequest) (*v1.AddResponse, error) {
	result := in.GetA() + in.GetB()
	return (&v1.AddResponse_builder{Result: &result}).Build(), nil
}

func (s *mockGRPCServer) Subtract(ctx context.Context, in *v1.SubtractRequest) (*v1.SubtractResponse, error) {
	result := in.GetA() - in.GetB()
	return (&v1.SubtractResponse_builder{Result: &result}).Build(), nil
}

func TestGRPCUpstream_Register_WithProtoCollection(t *testing.T) {
	// Create a temporary directory for proto files
	tempDir, err := os.MkdirTemp("", "proto-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a subdirectory for the proto files
	protoDir := filepath.Join(tempDir, "protos")
	err = os.Mkdir(protoDir, 0755)
	require.NoError(t, err)

	// Write a simple proto file
	protoContent := `
syntax = "proto3";
package test;
service TestService {
  rpc TestMethod (TestRequest) returns (TestResponse);
}
message TestRequest {
  string name = 1;
}
message TestResponse {
  string message = 1;
}
`
	protoFilePath := filepath.Join(protoDir, "test.proto")
	err = os.WriteFile(protoFilePath, []byte(protoContent), 0644)
	require.NoError(t, err)

	// Set up gRPC upstream
	poolManager := pool.NewManager()
	grpcUpstream := NewGRPCUpstream(poolManager)

	// Create service config with ProtoCollection
	protoCollection := (&configv1.ProtoCollection_builder{
		RootPath:       &tempDir,
		PathMatchRegex: proto.String(".*\\.proto"),
		IsRecursive:    proto.Bool(true),
	}).Build()

	grpcService := (&configv1.GrpcUpstreamService_builder{
		Address:          proto.String("localhost:50051"), // Dummy address
		ProtoCollections: []*configv1.ProtoCollection{protoCollection},
	}).Build()

	serviceConfig := (&configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("test-service"),
		GrpcService: grpcService,
	}).Build()

	toolManager := tool.NewToolManager(nil)

	// Register the service
	serviceKey, discoveredTools, err := grpcUpstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceKey)
	assert.NotNil(t, discoveredTools)

	// Verify that the tool was discovered and registered
	_, ok := toolManager.GetTool(fmt.Sprintf("%s.%s", serviceKey, "TestMethod"))
	assert.True(t, ok, "Tool TestMethod should be registered")

	// Check discovered tools
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "TestMethod", discoveredTools[0].GetName())
}

func TestGRPCUpstream_Register_WithReflection(t *testing.T) {
	// Start a test gRPC server with reflection
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	s := grpc.NewServer()
	v1.RegisterCalculatorServiceServer(s, &mockGRPCServer{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer s.Stop()

	// Set up gRPC upstream
	poolManager := pool.NewManager()
	grpcUpstream := NewGRPCUpstream(poolManager)

	// Create service config with reflection enabled
	grpcService := (&configv1.GrpcUpstreamService_builder{
		Address:       proto.String(lis.Addr().String()),
		UseReflection: proto.Bool(true),
	}).Build()

	serviceConfig := (&configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("test-service-reflection"),
		GrpcService: grpcService,
	}).Build()

	toolManager := tool.NewToolManager(nil)

	// Register the service
	serviceKey, discoveredTools, err := grpcUpstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceKey)
	assert.NotNil(t, discoveredTools)

	// Verify that the tool was discovered and registered
	_, ok := toolManager.GetTool(fmt.Sprintf("%s.%s", serviceKey, "Add"))
	assert.True(t, ok, "Tool Add should be registered")
}

func TestGRPCUpstream_Register_WithProtoDefinition(t *testing.T) {
	// Create a temporary directory for proto files
	tempDir, err := os.MkdirTemp("", "proto-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Write a simple proto file
	protoContent := `
syntax = "proto3";
package test;
service TestService {
  rpc TestMethod (TestRequest) returns (TestResponse);
}
message TestRequest {
  string name = 1;
}
message TestResponse {
  string message = 1;
}
`
	protoFilePath := filepath.Join(tempDir, "test.proto")
	err = os.WriteFile(protoFilePath, []byte(protoContent), 0644)
	require.NoError(t, err)

	// Set up gRPC upstream
	poolManager := pool.NewManager()
	grpcUpstream := NewGRPCUpstream(poolManager)

	// Create service config with ProtoDefinition
	protoFile := (&configv1.ProtoFile_builder{
		FileName:    proto.String("test.proto"),
		FileContent: proto.String(protoContent),
	}).Build()
	protoDefinition := (&configv1.ProtoDefinition_builder{
		ProtoFile: protoFile,
	}).Build()

	grpcService := (&configv1.GrpcUpstreamService_builder{
		Address:          proto.String("localhost:50051"), // Dummy address
		ProtoDefinitions: []*configv1.ProtoDefinition{protoDefinition},
	}).Build()

	serviceConfig := (&configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("test-service-proto-definition"),
		GrpcService: grpcService,
	}).Build()

	toolManager := tool.NewToolManager(nil)

	// Register the service
	serviceKey, discoveredTools, err := grpcUpstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceKey)
	assert.NotNil(t, discoveredTools)

	// Verify that the tool was discovered and registered
	_, ok := toolManager.GetTool(fmt.Sprintf("%s.%s", serviceKey, "TestMethod"))
	assert.True(t, ok, "Tool TestMethod should be registered")

	// Check discovered tools
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "TestMethod", discoveredTools[0].GetName())
}
