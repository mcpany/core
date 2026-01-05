// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package factory

import (
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/upstream/command"
	"github.com/mcpany/core/server/pkg/upstream/grpc"
	"github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/mcpany/core/server/pkg/upstream/mcp"
	"github.com/mcpany/core/server/pkg/upstream/openapi"
	"github.com/mcpany/core/server/pkg/upstream/webrtc"
	"github.com/mcpany/core/server/pkg/upstream/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpstreamServiceFactory(t *testing.T) {
	t.Run("with a valid pool manager", func(t *testing.T) {
		pm := pool.NewManager()
		f := NewUpstreamServiceFactory(pm)
		assert.NotNil(t, f)
		impl, ok := f.(*UpstreamServiceFactory)
		assert.True(t, ok)
		assert.Equal(t, pm, impl.poolManager)
	})

	t.Run("with a nil pool manager", func(t *testing.T) {
		f := NewUpstreamServiceFactory(nil)
		assert.NotNil(t, f)
		impl, ok := f.(*UpstreamServiceFactory)
		assert.True(t, ok)
		assert.Nil(t, impl.poolManager)
	})
}

func TestUpstreamServiceFactory_NewUpstream(t *testing.T) {
	pm := pool.NewManager()
	f := NewUpstreamServiceFactory(pm)

	grpcConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{},
		},
	}

	httpConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{},
		},
	}

	openapiConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: &configv1.OpenapiUpstreamService{},
		},
	}

	mcpConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{},
		},
	}

	commandLineConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{},
		},
	}

	websocketConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
			WebsocketService: &configv1.WebsocketUpstreamService{},
		},
	}

	webrtcConfig := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
			WebrtcService: &configv1.WebrtcUpstreamService{},
		},
	}

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
			name:        "Unknown Service",
			config:      &configv1.UpstreamServiceConfig{},
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
