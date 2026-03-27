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
// Summary. Represents a GraphQLProvider.
type GraphQLProvider struct {
	Endpoint string // e.g., "http://localhost:8080/graphql"
}

// Name provides name functionality.
//
// Summary: Name.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (p *GraphQLProvider) Name() string {
	return "graphql"
}

// Discover provides discover functionality.
//
// Summary: Discover.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
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
