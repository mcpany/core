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
	"os"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

func findMethodDescriptorForZZTest(t *testing.T, serviceName, methodName string) protoreflect.MethodDescriptor {
	t.Helper()
	b, err := os.ReadFile("../../build/all.protoset")
	require.NoError(t, err, "Failed to read protoset file. Ensure 'make gen' has been run.")

	fds := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(b, fds)
	require.NoError(t, err, "Failed to unmarshal protoset file")

	files, err := protodesc.NewFiles(fds)
	require.NoError(t, err)

	var methodDesc protoreflect.MethodDescriptor
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			if string(service.Name()) == serviceName {
				method := service.Methods().ByName(protoreflect.Name(methodName))
				if method != nil {
					methodDesc = method
					return false // stop iterating
				}
			}
		}
		return true
	})

	require.NotNil(t, methodDesc, "method %s not found in service %s", methodName, serviceName)
	return methodDesc
}

func TestToolMethods(t *testing.T) {
	mcpTool := &v1.Tool{}
	cacheConfig := configv1.CacheConfig_builder{
		Ttl: durationpb.New(10),
	}.Build()

	t.Run("HTTPTool", func(t *testing.T) {
		poolManager := pool.NewManager()
		p, err := pool.New(func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: nil}, nil
		}, 1, 1, 0, true)
		require.NoError(t, err)
		poolManager.Register("test-service", p)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, configv1.HttpCallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		assert.Equal(t, mcpTool, httpTool.Tool())
		assert.Equal(t, cacheConfig, httpTool.GetCacheConfig())
	})

	t.Run("GRPCTool", func(t *testing.T) {
		poolManager := pool.NewManager()
		p, err := pool.New(func(ctx context.Context) (*client.GrpcClientWrapper, error) {
			return &client.GrpcClientWrapper{}, nil
		}, 1, 1, 0, true)
		require.NoError(t, err)
		poolManager.Register("test-service", p)

		method := findMethodDescriptorForZZTest(t, "WeatherService", "GetWeather")
		grpcTool := tool.NewGRPCTool(mcpTool, poolManager, "test-service", method, configv1.GrpcCallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		assert.Equal(t, mcpTool, grpcTool.Tool())
		assert.Equal(t, cacheConfig, grpcTool.GetCacheConfig())
	})

	t.Run("MCPTool", func(t *testing.T) {
		mcpToolObject := tool.NewMCPTool(mcpTool, nil, configv1.MCPCallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		assert.Equal(t, mcpTool, mcpToolObject.Tool())
		assert.Equal(t, cacheConfig, mcpToolObject.GetCacheConfig())
	})

	t.Run("OpenAPITool", func(t *testing.T) {
		openAPITool := tool.NewOpenAPITool(mcpTool, nil, nil, "", "", nil, configv1.OpenAPICallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		assert.Equal(t, mcpTool, openAPITool.Tool())
		assert.Equal(t, cacheConfig, openAPITool.GetCacheConfig())
	})

	t.Run("CommandTool", func(t *testing.T) {
		commandTool := tool.NewCommandTool(mcpTool, &configv1.CommandLineUpstreamService{}, configv1.CommandLineCallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		assert.Equal(t, mcpTool, commandTool.Tool())
		assert.Equal(t, cacheConfig, commandTool.GetCacheConfig())
	})

	t.Run("WebrtcTool", func(t *testing.T) {
		webrtcTool, err := tool.NewWebrtcTool(mcpTool, nil, "test-service", nil, configv1.WebrtcCallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		require.NoError(t, err)
		assert.Equal(t, mcpTool, webrtcTool.Tool())
		assert.Equal(t, cacheConfig, webrtcTool.GetCacheConfig())
	})

	t.Run("WebsocketTool", func(t *testing.T) {
		websocketTool := tool.NewWebsocketTool(mcpTool, nil, "test-service", nil, configv1.WebsocketCallDefinition_builder{
			Cache: cacheConfig,
		}.Build())
		assert.Equal(t, mcpTool, websocketTool.Tool())
		assert.Equal(t, cacheConfig, websocketTool.GetCacheConfig())
	})
}
