// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package factory

import (
	"fmt"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/upstream/command"
	"github.com/mcpany/core/pkg/upstream/grpc"
	"github.com/mcpany/core/pkg/upstream/http"
	"github.com/mcpany/core/pkg/upstream/mcp"
	"github.com/mcpany/core/pkg/upstream/openapi"
	"github.com/mcpany/core/pkg/upstream/webrtc"
	"github.com/mcpany/core/pkg/upstream/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Factory defines the interface for a factory that creates upstream service
// instances.
type Factory interface {
	// NewUpstream creates a new upstream service instance based on the provided
	// configuration.
	NewUpstream(config *configv1.UpstreamServiceConfig) (upstream.Upstream, error)
}

// UpstreamServiceFactory is a concrete implementation of the Factory interface.
// It creates different types of upstream services based on the service
// configuration.
type UpstreamServiceFactory struct {
	poolManager *pool.Manager
}

// NewUpstreamServiceFactory creates a new UpstreamServiceFactory.
//
// poolManager is the connection pool manager used by upstreams that require
// connection pooling (e.g., gRPC, HTTP, WebSocket).
func NewUpstreamServiceFactory(poolManager *pool.Manager) Factory {
	return &UpstreamServiceFactory{
		poolManager: poolManager,
	}
}

// NewUpstream creates and returns an appropriate upstream.Upstream implementation
// based on the type of service specified in the configuration.
//
// config is the configuration for the upstream service.
// It returns a new upstream service instance or an error if the service type is
// unknown.
func (f *UpstreamServiceFactory) NewUpstream(config *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	if config == nil {
		return nil, fmt.Errorf("upstream service config cannot be nil")
	}
	switch config.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_GrpcService_case:
		return grpc.NewGRPCUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_HttpService_case:
		return http.NewHTTPUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		return openapi.NewOpenAPIUpstream(), nil
	case configv1.UpstreamServiceConfig_McpService_case:
		return mcp.NewMCPUpstream(), nil
	case configv1.UpstreamServiceConfig_CommandLineService_case:
		return command.NewCommandUpstream(), nil
	case configv1.UpstreamServiceConfig_WebsocketService_case:
		return websocket.NewWebsocketUpstream(f.poolManager), nil
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		return webrtc.NewWebrtcUpstream(f.poolManager), nil
	default:
		return nil, fmt.Errorf("unknown service config type: %T", config.WhichServiceConfig())
	}
}
