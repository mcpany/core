// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package factory provides upstream factory functionality.
package factory

import (
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/upstream"
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
)

// Factory factory represents a factory.
//
// Summary: Factory represents a factory.
type Factory interface {
	// NewUpstream creates a new upstream service instance based on the provided
	// configuration.
	//
	// Summary: Creates a new upstream service.
	//
	// Parameters: - None.
	//   - config: *configv1.UpstreamServiceConfig. The upstream service configuration.
	//
	// Returns: - None.
	//   - upstream.Upstream: The created upstream service.
	//   - error: An error if creation fails.
	NewUpstream(config *configv1.UpstreamServiceConfig) (upstream.Upstream, error)
}

// UpstreamServiceFactory upstreamServiceFactory represents a upstream service factory.
//
// Summary: UpstreamServiceFactory represents a upstream service factory.
type UpstreamServiceFactory struct {
	poolManager    *pool.Manager
	globalSettings *configv1.GlobalSettings
}

// NewUpstreamServiceFactory creates a new UpstreamServiceFactory.
//
// Summary: Creates a new UpstreamServiceFactory.
//
// Parameters: - None.
//   - poolManager: *pool.Manager. The connection pool manager used by upstreams that require.
//     connection pooling (e.g., gRPC, HTTP, WebSocket).
//   - globalSettings: *configv1.GlobalSettings. The global configuration settings.
//
// Returns: - None.
//   - Factory: A new Factory instance.
func NewUpstreamServiceFactory(poolManager *pool.Manager, globalSettings *configv1.GlobalSettings) Factory {
	return &UpstreamServiceFactory{
		poolManager:    poolManager,
		globalSettings: globalSettings,
	}
}

// NewUpstream creates and returns an appropriate upstream.Upstream implementation
// based on the type of service specified in the configuration.
//
// Summary: Creates a new upstream service based on configuration.
//
// Parameters: - None.
//   - config: *configv1.UpstreamServiceConfig. The configuration for the upstream service.
//
// Returns: - None.
//   - upstream.Upstream: A new upstream service instance.
//   - error: An error if the service type is unknown.
func (f *UpstreamServiceFactory) NewUpstream(config *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	if config == nil {
		return nil, fmt.Errorf("upstream service config cannot be nil")
	}
	switch config.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_GrpcService_case:
		return grpc.NewUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_HttpService_case:
		return http.NewUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		return openapi.NewOpenAPIUpstream(), nil
	case configv1.UpstreamServiceConfig_McpService_case:
		return mcp.NewUpstream(f.globalSettings), nil
	case configv1.UpstreamServiceConfig_CommandLineService_case:
		return command.NewUpstream(), nil
	case configv1.UpstreamServiceConfig_WebsocketService_case:
		return websocket.NewUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		return webrtc.NewUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_GraphqlService_case:
		return graphql.NewGraphQLUpstream(), nil
	case configv1.UpstreamServiceConfig_SqlService_case:
		return sql.NewUpstream(), nil
	case configv1.UpstreamServiceConfig_FilesystemService_case:
		return filesystem.NewUpstream(), nil
	case configv1.UpstreamServiceConfig_VectorService_case:
		return vector.NewUpstream(), nil
	default:
		return nil, fmt.Errorf("unknown service config type: %T", config.WhichServiceConfig())
	}
}
