package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/grpc/protobufparser"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGRPCUpstream_createAndRegisterGRPCToolsFromConfig(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	serviceID := "test-service"
	tm.AddServiceInfo(serviceID, &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("test-tool"),
						CallId: proto.String("test-call"),
					}.Build(),
				},
				Calls: map[string]*configv1.GrpcCallDefinition{
					"test-call": configv1.GrpcCallDefinition_builder{}.Build(),
				},
			}.Build(),
		}.Build(),
	})

	t.Run("nil fds", func(t *testing.T) {
		tools, err := upstream.(*Upstream).createAndRegisterGRPCToolsFromConfig(
			context.Background(),
			serviceID,
			tm,
			nil,
			false,
			nil,
		)
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
		_, err := upstream.(*Upstream).createAndRegisterGRPCToolsFromConfig(
			context.Background(),
			serviceID,
			tm,
			nil,
			false,
			fds,
		)
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
					Tools: []*configv1.ToolDefinition{
						configv1.ToolDefinition_builder{
							Name:   proto.String("test-tool"),
							CallId: proto.String("non-existent-call"),
						}.Build(),
					},
					Calls: map[string]*configv1.GrpcCallDefinition{
						"test-call": configv1.GrpcCallDefinition_builder{}.Build(),
					},
				}.Build(),
			}.Build(),
		})

		tools, err := upstream.(*Upstream).createAndRegisterGRPCToolsFromConfig(
			context.Background(),
			serviceID,
			tm,
			nil,
			false,
			fds,
		)
		require.NoError(t, err)
		assert.Empty(t, tools)
	})

	t.Run("successful tool registration", func(t *testing.T) {
		server, addr := startMockServer(t)
		defer server.Stop()
		ctx := context.Background()
		fds, err := protobufparser.ParseProtoByReflection(ctx, addr)
		require.NoError(t, err)

		serviceConfig := &configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("GetWeather"),
						CallId: proto.String("get-weather-call"),
					}.Build(),
				},
				Calls: map[string]*configv1.GrpcCallDefinition{
					"get-weather-call": configv1.GrpcCallDefinition_builder{
						Service: proto.String("examples.weather.v1.WeatherService"),
						Method:  proto.String("GetWeather"),
					}.Build(),
				},
			}.Build(),
		}

		tm.AddServiceInfo(serviceID, &tool.ServiceInfo{Config: serviceConfig.Build()})

		discoveredTools, err := upstream.(*Upstream).createAndRegisterGRPCToolsFromConfig(
			context.Background(), serviceID, tm, nil, false, fds,
		)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 1)
		assert.Len(t, tm.ListTools(), 1)
		assert.Equal(t, "GetWeather", discoveredTools[0].GetName())
	})
}

func TestGRPCUpstream_createAndRegisterPromptsFromConfig(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	promptManager := prompt.NewManager()
	serviceConfig := &configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GrpcService: configv1.GrpcUpstreamService_builder{
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name: proto.String("test-prompt"),
				}.Build(),
			},
		}.Build(),
	}
	tm.AddServiceInfo("test-service", &tool.ServiceInfo{Config: serviceConfig.Build()})
	upstream.(*Upstream).toolManager = tm

	err := upstream.(*Upstream).createAndRegisterPromptsFromConfig(context.Background(), "test-service", promptManager, false)
	require.NoError(t, err)
	_, ok := promptManager.GetPrompt("test-service.test-prompt")
	assert.True(t, ok)
}
