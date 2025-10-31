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

package factory

import (
	"testing"

	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/upstream/command"
	"github.com/mcpxy/core/pkg/upstream/grpc"
	"github.com/mcpxy/core/pkg/upstream/http"
	"github.com/mcpxy/core/pkg/upstream/mcp"
	"github.com/mcpxy/core/pkg/upstream/openapi"
	"github.com/mcpxy/core/pkg/upstream/webrtc"
	"github.com/mcpxy/core/pkg/upstream/websocket"
	configv1 "github.com/mcpxy/core/proto/config/v1"
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

	grpcConfig := &configv1.UpstreamServiceConfig{}
	grpcConfig.SetGrpcService(&configv1.GrpcUpstreamService{})

	httpConfig := &configv1.UpstreamServiceConfig{}
	httpConfig.SetHttpService(&configv1.HttpUpstreamService{})

	openapiConfig := &configv1.UpstreamServiceConfig{}
	openapiConfig.SetOpenapiService(&configv1.OpenapiUpstreamService{})

	mcpConfig := &configv1.UpstreamServiceConfig{}
	mcpConfig.SetMcpService(&configv1.McpUpstreamService{})

	commandLineConfig := &configv1.UpstreamServiceConfig{}
	commandLineConfig.SetCommandLineService(&configv1.CommandLineUpstreamService{})

	websocketConfig := &configv1.UpstreamServiceConfig{}
	websocketConfig.SetWebsocketService(&configv1.WebsocketUpstreamService{})

	webrtcConfig := &configv1.UpstreamServiceConfig{}
	webrtcConfig.SetWebrtcService(&configv1.WebrtcUpstreamService{})

	testCases := []struct {
		name        string
		config      *configv1.UpstreamServiceConfig
		expectedTyp interface{}
		expectError bool
	}{
		{
			name:        "gRPC Service",
			config:      grpcConfig,
			expectedTyp: &grpc.GRPCUpstream{},
		},
		{
			name:        "HTTP Service",
			config:      httpConfig,
			expectedTyp: &http.HTTPUpstream{},
		},
		{
			name:        "OpenAPI Service",
			config:      openapiConfig,
			expectedTyp: &openapi.OpenAPIUpstream{},
		},
		{
			name:        "MCP Service",
			config:      mcpConfig,
			expectedTyp: &mcp.MCPUpstream{},
		},
		{
			name:        "Command Line Service",
			config:      commandLineConfig,
			expectedTyp: &command.CommandUpstream{},
		},
		{
			name:        "Websocket Service",
			config:      websocketConfig,
			expectedTyp: &websocket.WebsocketUpstream{},
		},
		{
			name:        "WebRTC Service",
			config:      webrtcConfig,
			expectedTyp: &webrtc.WebrtcUpstream{},
		},
		{
			name:        "Unknown Service",
			config:      &configv1.UpstreamServiceConfig{},
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
