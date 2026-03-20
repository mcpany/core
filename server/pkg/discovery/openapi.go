// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: OpenAPIProvider discovers services via OpenAPI specifications.
//
// Side Effects:
//   - None.
package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

type OpenAPIProvider struct {
	Endpoint string // e.g., "http://localhost:8080/openapi.json"
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

func (p *OpenAPIProvider) Name() string {
	return "openapi"
}

func (p *OpenAPIProvider) Discover(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	if p.Endpoint == "" {
		return nil, nil
	}

	// Create a dynamic configuration for the OpenAPI service
	return []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{
			Name:    proto.String("Auto-discovered OpenAPI"),
			Version: proto.String("v1"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				Address: proto.String(p.Endpoint),
			}.Build(),
			Tags: []string{"openapi", "auto-discovered"},
		}.Build(),
	}, nil
}
