// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package factory provides upstream factory functionality.
package factory

import (
	"fmt"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/upstream/command"
	"github.com/mcpany/core/pkg/upstream/graphql"
	"github.com/mcpany/core/pkg/upstream/grpc"
	"github.com/mcpany/core/pkg/upstream/http"
	"github.com/mcpany/core/pkg/upstream/mcp"
	"github.com/mcpany/core/pkg/upstream/openapi"
	"github.com/mcpany/core/pkg/upstream/sql"
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

	// ShutdownUpstream shuts down the upstream service with the given name.
	ShutdownUpstream(serviceName string) error
}

// UpstreamServiceFactory is a concrete implementation of the Factory interface.
// It creates different types of upstream services based on the service
// configuration.
type UpstreamServiceFactory struct {
	poolManager *pool.Manager
}

// NewUpstreamServiceFactory creates a new UpstreamServiceFactory.
//
// Parameters:
//   poolManager: The connection pool manager used by upstreams that require
//   connection pooling (e.g., gRPC, HTTP, WebSocket).
//
// Returns:
//   Factory: A new Factory instance.
func NewUpstreamServiceFactory(poolManager *pool.Manager) Factory {
	return &UpstreamServiceFactory{
		poolManager: poolManager,
	}
}

// ShutdownUpstream shuts down the upstream service with the given name by deregistering
// its connection pool.
func (f *UpstreamServiceFactory) ShutdownUpstream(serviceName string) error {
	f.poolManager.Deregister(serviceName)
	return nil
}

// NewUpstream creates and returns an appropriate upstream.Upstream implementation
// based on the type of service specified in the configuration.
//
// Parameters:
//   config: The configuration for the upstream service.
//
// Returns:
//   upstream.Upstream: A new upstream service instance.
//   error: An error if the service type is unknown.
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
		return mcp.NewUpstream(), nil
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
	default:
		return nil, fmt.Errorf("unknown service config type: %T", config.WhichServiceConfig())
	}
}
