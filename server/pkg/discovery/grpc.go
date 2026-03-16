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
// Summary: GRPCProvider discovers services via gRPC reflection.
//
// Summary: GRPCProvider discovers services via gRPC reflection.
type GRPCProvider struct {
	Endpoint string // e.g., "localhost:50051"
// Name returns the name of the provider.
//
// Summary: Name returns the name of the provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.

// Name returns the name of the provider.
//
// Summary: Name returns the name of the provider.
// Discover attempts to find services and return their configurations.
//
// Summary: Discover attempts to find services and return their configurations.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (p *GRPCProvider) Name() string {
	return "grpc"
}

// Discover attempts to find services and return their configurations.
//
// Summary: Discover attempts to find services and return their configurations.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
