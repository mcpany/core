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
// Summary: Represents a GraphQLProvider.
type GraphQLProvider struct {
	Endpoint string // e.g., "http://localhost:8080/graphql"
}

// Name name name.
//
// Summary: Name name.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - string: The result.
func (p *GraphQLProvider) Name() string {
	return "graphql"
}

// Discover discover discover.
//
// Summary: Discover discover.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//
// Returns: - None.
//   - []*configv1.UpstreamServiceConfig: The result.
//   - error: An error if the operation fails.
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
