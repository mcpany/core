package factory

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/upstream/command"
	"github.com/mcpany/core/server/pkg/upstream/filesystem"
	"github.com/mcpany/core/server/pkg/upstream/graphql"
	"github.com/mcpany/core/server/pkg/upstream/grpc"
	"github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/mcpany/core/server/pkg/upstream/mcp"
	"github.com/mcpany/core/server/pkg/upstream/openapi"
	"github.com/mcpany/core/server/pkg/upstream/sql"
	"github.com/mcpany/core/server/pkg/upstream/vector"
	"github.com/mcpany/core/server/pkg/upstream/webrtc"
	"github.com/mcpany/core/server/pkg/upstream/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpstreamServiceFactory(t *testing.T) {
	t.Run("with a valid pool manager", func(t *testing.T) {
		pm := pool.NewManager()
		f := NewUpstreamServiceFactory(pm, nil)
		assert.NotNil(t, f)
		impl, ok := f.(*UpstreamServiceFactory)
		assert.True(t, ok)
		assert.Equal(t, pm, impl.poolManager)
	})

	t.Run("with a nil pool manager", func(t *testing.T) {
		f := NewUpstreamServiceFactory(nil, nil)
		assert.NotNil(t, f)
		impl, ok := f.(*UpstreamServiceFactory)
		assert.True(t, ok)
		assert.Nil(t, impl.poolManager)
	})
}

func TestUpstreamServiceFactory_NewUpstream(t *testing.T) {
	pm := pool.NewManager()
	f := NewUpstreamServiceFactory(pm, nil)

	grpcConfig := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{}.Build(),
	}.Build()

	httpConfig := configv1.UpstreamServiceConfig_builder{
		HttpService: configv1.HttpUpstreamService_builder{}.Build(),
	}.Build()

	openapiConfig := configv1.UpstreamServiceConfig_builder{
		OpenapiService: configv1.OpenapiUpstreamService_builder{}.Build(),
	}.Build()

	mcpConfig := configv1.UpstreamServiceConfig_builder{
		McpService: configv1.McpUpstreamService_builder{}.Build(),
	}.Build()

	commandLineConfig := configv1.UpstreamServiceConfig_builder{
		CommandLineService: configv1.CommandLineUpstreamService_builder{}.Build(),
	}.Build()

	websocketConfig := configv1.UpstreamServiceConfig_builder{
		WebsocketService: configv1.WebsocketUpstreamService_builder{}.Build(),
	}.Build()

	webrtcConfig := configv1.UpstreamServiceConfig_builder{
		WebrtcService: configv1.WebrtcUpstreamService_builder{}.Build(),
	}.Build()

	graphqlConfig := configv1.UpstreamServiceConfig_builder{
		GraphqlService: configv1.GraphQLUpstreamService_builder{}.Build(),
	}.Build()

	sqlConfig := configv1.UpstreamServiceConfig_builder{
		SqlService: configv1.SqlUpstreamService_builder{}.Build(),
	}.Build()

	filesystemConfig := configv1.UpstreamServiceConfig_builder{
		FilesystemService: configv1.FilesystemUpstreamService_builder{}.Build(),
	}.Build()

	vectorConfig := configv1.UpstreamServiceConfig_builder{
		VectorService: configv1.VectorUpstreamService_builder{}.Build(),
	}.Build()

	testCases := []struct {
		name        string
		config      *configv1.UpstreamServiceConfig
		expectedTyp interface{}
		expectError bool
	}{
		{
			name:        "gRPC Service",
			config:      grpcConfig,
			expectedTyp: &grpc.Upstream{},
		},
		{
			name:        "HTTP Service",
			config:      httpConfig,
			expectedTyp: &http.Upstream{},
		},
		{
			name:        "OpenAPI Service",
			config:      openapiConfig,
			expectedTyp: &openapi.OpenAPIUpstream{},
		},
		{
			name:        "MCP Service",
			config:      mcpConfig,
			expectedTyp: &mcp.Upstream{},
		},
		{
			name:        "Command Line Service",
			config:      commandLineConfig,
			expectedTyp: &command.Upstream{},
		},
		{
			name:        "Websocket Service",
			config:      websocketConfig,
			expectedTyp: &websocket.Upstream{},
		},
		{
			name:        "WebRTC Service",
			config:      webrtcConfig,
			expectedTyp: &webrtc.Upstream{},
		},
		{
			name:        "GraphQL Service",
			config:      graphqlConfig,
			expectedTyp: &graphql.Upstream{},
		},
		{
			name:        "SQL Service",
			config:      sqlConfig,
			expectedTyp: &sql.Upstream{},
		},
		{
			name:        "Filesystem Service",
			config:      filesystemConfig,
			expectedTyp: &filesystem.Upstream{},
		},
		{
			name:        "Vector Service",
			config:      vectorConfig,
			expectedTyp: &vector.Upstream{},
		},
		{
			name:        "Unknown Service",
			config:      configv1.UpstreamServiceConfig_builder{}.Build(),
			expectError: true,
		},
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := f.NewUpstream(tc.config)
			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, u)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, u)
				assert.IsType(t, tc.expectedTyp, u)
			}
		})
	}
}
