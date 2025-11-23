// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/grpc/protobufparser"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGRPCUpstream_createAndRegisterGRPCTools_DisabledTool(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := NewMockToolManager()
	tm.AddServiceInfo("test-service", &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:    proto.String("test-tool"),
						Disable: proto.Bool(true),
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

	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name: proto.String("test.proto"),
			},
		},
	}

	tools, err := upstream.(*GRPCUpstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, nil, false, fds)
	require.NoError(t, err)
	assert.Empty(t, tools)
}

func TestGRPCUpstream_createAndRegisterGRPCTools_MissingMethodDescriptor(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := NewMockToolManager()
	tm.AddServiceInfo("test-service", &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{}.Build(),
		}.Build(),
	})

	parsedData := &protobufparser.ParsedMcpAnnotations{
		Tools: []protobufparser.McpTool{
			{Name: "test-tool", FullMethodName: "nonexistent.Service/Method"},
		},
	}
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name: proto.String("test.proto"),
			},
		},
	}

	tools, err := upstream.(*GRPCUpstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, nil, false, fds)
	require.NoError(t, err)
	assert.Empty(t, tools)
}

func TestGRPCUpstream_createAndRegisterGRPCTools_DynamicResourceMissingTool(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := tool.NewToolManager(nil)
	rm := resource.NewResourceManager()
	tm.AddServiceInfo("test-service", &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Resources: []*configv1.ResourceDefinition{
					configv1.ResourceDefinition_builder{
						Name: proto.String("test-resource"),
						Dynamic: configv1.DynamicResource_builder{
							GrpcCall: configv1.GrpcCallDefinition_builder{
								Id: proto.String("missing-tool"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	})

	parsedData := &protobufparser.ParsedMcpAnnotations{}
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name: proto.String("test.proto"),
			},
		},
	}
	_, err := upstream.(*GRPCUpstream).createAndRegisterGRPCTools(context.Background(), "test-service", parsedData, tm, rm, false, fds)
	require.NoError(t, err)
	assert.Empty(t, rm.ListResources())
}

func TestGRPCUpstream_createAndRegisterGRPCToolsFromDescriptors_DisabledTool(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := NewMockToolManager()
	tm.AddServiceInfo("test-service", &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:    proto.String("TestMethod"),
						Disable: proto.Bool(true),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	})

	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("test.proto"),
				Package: proto.String("test"),
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("TestRequest"),
					},
					{
						Name: proto.String("TestResponse"),
					},
				},
				Service: []*descriptorpb.ServiceDescriptorProto{
					{
						Name: proto.String("TestService"),
						Method: []*descriptorpb.MethodDescriptorProto{
							{
								Name:       proto.String("TestMethod"),
								InputType:  proto.String(".test.TestRequest"),
								OutputType: proto.String(".test.TestResponse"),
							},
						},
					},
				},
			},
		},
	}

	tools, err := upstream.(*GRPCUpstream).createAndRegisterGRPCToolsFromDescriptors(context.Background(), "test-service", tm, nil, false, fds)
	require.NoError(t, err)
	assert.Empty(t, tools)
}
