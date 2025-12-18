// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGRPCUpstream_Register_AutoDiscover_False_NoAnnotations_DiscoversNothing(t *testing.T) {
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

	// AutoDiscoverTool is NOT set (defaults to false)
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service-no-autodiscover")
	serviceConfig.SetGrpcService(grpcService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	assert.Empty(t, discoveredTools, "Should not discover tools when AutoDiscoverTool is false and no annotations are present")
}

func TestGRPCUpstream_Register_AutoDiscover_True_NoAnnotations_DiscoversEverything(t *testing.T) {
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

	// AutoDiscoverTool IS set
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service-autodiscover")
	serviceConfig.SetGrpcService(grpcService)
	serviceConfig.AutoDiscoverTool = proto.Bool(true)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	assert.NotEmpty(t, discoveredTools, "Should discover tools when AutoDiscoverTool is true")
	if len(discoveredTools) > 0 {
		assert.Equal(t, "GetData", discoveredTools[0].GetName())
	}
}
