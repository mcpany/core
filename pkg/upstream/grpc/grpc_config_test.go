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

package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/grpc/protobufparser"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGRPCUpstream_createAndRegisterGRPCToolsFromConfig(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewGRPCUpstream(poolManager)
	tm := NewMockToolManager()

	serviceID := "test-service"
	tm.AddServiceInfo(serviceID, &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Tools: []*configv1.GrpcToolDefinition{
					configv1.GrpcToolDefinition_builder{
						Schema: configv1.ToolSchema_builder{
							Name: proto.String("test-tool"),
						}.Build(),
						CallId: proto.String("test-call"),
					}.Build(),
				},
				Calls: map[string]*configv1.GrpcCallDefinition{
					"test-call": configv1.GrpcCallDefinition_builder{
						Schema: configv1.ToolSchema_builder{
							Name: proto.String("test-tool"),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	})

	t.Run("nil fds", func(t *testing.T) {
		tools, err := upstream.(*GRPCUpstream).createAndRegisterGRPCToolsFromConfig(context.Background(), serviceID, tm, nil, false, nil)
		require.NoError(t, err)
		assert.Nil(t, tools)
	})

	t.Run("bad file descriptor set", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				{
					Name:       proto.String("test.proto"),
					Dependency: []string{"nonexistent.proto"},
				},
			},
		}
		_, err := upstream.(*GRPCUpstream).createAndRegisterGRPCToolsFromConfig(context.Background(), serviceID, tm, nil, false, fds)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create protodesc files")
	})

	t.Run("call definition not found", func(t *testing.T) {
		server, addr := startMockServer(t)
		defer server.Stop()
		ctx := context.Background()
		fds, err := protobufparser.ParseProtoByReflection(ctx, addr)
		require.NoError(t, err)

		tm.AddServiceInfo(serviceID, &tool.ServiceInfo{
			Config: configv1.UpstreamServiceConfig_builder{
				GrpcService: configv1.GrpcUpstreamService_builder{
					Tools: []*configv1.GrpcToolDefinition{
						configv1.GrpcToolDefinition_builder{
							Schema: configv1.ToolSchema_builder{
								Name: proto.String("test-tool"),
							}.Build(),
							CallId: proto.String("non-existent-call"),
						}.Build(),
					},
					Calls: map[string]*configv1.GrpcCallDefinition{
						"test-call": configv1.GrpcCallDefinition_builder{
							Schema: configv1.ToolSchema_builder{
								Name: proto.String("test-tool"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			}.Build(),
		})

		tools, err := upstream.(*GRPCUpstream).createAndRegisterGRPCToolsFromConfig(context.Background(), serviceID, tm, nil, false, fds)
		require.NoError(t, err)
		assert.Empty(t, tools)
	})
}
