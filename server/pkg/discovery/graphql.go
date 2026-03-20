// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: GraphQLProvider discovers services via GraphQL introspection.
//
// Side Effects:
//   - None.
package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

type GraphQLProvider struct {
	Endpoint string // e.g., "http://localhost:8080/graphql"
	// Summary: Name returns the name of the provider.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - unnamed (string): description
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	// Summary: Discover attempts to find services and return their configurations.
	//
	// Parameters:
	//   - _ (context.Context): description
	//
	// Returns:
	//   - unnamed (array/slice): description
	//   - unnamed (error): description
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
}

func (p *GraphQLProvider) Name() string {
	return "graphql"
}

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
