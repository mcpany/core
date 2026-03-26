// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// GraphQLProvider discovers services via GraphQL introspection.
//
// Summary: Discovery provider that uses GraphQL introspection.
type GraphQLProvider struct {
	Endpoint string // e.g., "http://localhost:8080/graphql"
}

// Name returns the name of the provider.
//
// Summary: Returns the canonical name for this GraphQL discovery provider.
//
// Returns:
//   - string: The name of the provider.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *GraphQLProvider) Name() string {
	return "graphql"
}

// Discover attempts to find services and return their configurations.
//
// Summary: Generates a dynamic service configuration for a GraphQL upstream by targeting the configured endpoint for introspection.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The discovered GraphQL service configurations.
//   - error: An error if discovery fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (p *GraphQLProvider) Discover(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	if p.Endpoint == "" {
		return nil, nil
	}

	// Create a dynamic configuration for the GraphQL service
	return []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{
			Name:    proto.String("Auto-discovered GraphQL"),
			Version: proto.String("v1"),
			GraphqlService: configv1.GraphQLUpstreamService_builder{
				Address: proto.String(p.Endpoint),
			}.Build(),
			Tags: []string{"graphql", "auto-discovered"},
		}.Build(),
	}, nil
}
