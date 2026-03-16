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
// Summary: GraphQLProvider discovers services via GraphQL introspection.
type GraphQLProvider struct {
	Endpoint string // e.g., "http://localhost:8080/graphql"
}

// Name returns the name of the provider.
//
// Summary: Name returns the name of the provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The string result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (p *GraphQLProvider) Name() string {
	return "graphql"
}

// Discover attempts to find services and return their configurations.
//
// Summary: Discover attempts to find services and return their configurations.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The []*configv1.UpstreamServiceConfig result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
