// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package discovery implements the discovery subsystem.
package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// GRPCProvider discovers services via gRPC reflection.
type GRPCProvider struct {
	Endpoint string // e.g., "localhost:50051"
}

// Name handles name.
//
// Parameters:
//   - None
//
// Returns:
//   - string: The generated or retrieved entity.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Name returns the name of the provider.
func (p *GRPCProvider) Name() string {
	return "grpc"
}

// Discover handles discover.
//
// Parameters:
//   - _ (context.Context): Reserved parameter, currently unused.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A slice containing the requested elements.
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
// Discover attempts to find services and return their configurations.
func (p *GRPCProvider) Discover(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	if p.Endpoint == "" {
		return nil, nil
	}

	// Create a dynamic configuration for the gRPC service
	return []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{
			Name:    proto.String("Auto-discovered gRPC"),
			Version: proto.String("v1"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address:       proto.String(p.Endpoint),
				UseReflection: proto.Bool(true),
			}.Build(),
			Tags: []string{"grpc", "auto-discovered"},
		}.Build(),
	}, nil
}
