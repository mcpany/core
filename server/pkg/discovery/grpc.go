// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// GRPCProvider discovers services via gRPC reflection.
//
// Summary: Represents a GRPCProvider.
type GRPCProvider struct {
	Endpoint string // e.g., "localhost:50051"
}

// Name name name.
//
// Summary: Name name.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *GRPCProvider) Name() string {
	return "grpc"
}

// Discover discover discover.
//
// Summary: Discover discover.
//
// Parameters:
//   - _ (context.Context): Unused parameter.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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
