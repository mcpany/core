// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// OpenAPIProvider discovers services via OpenAPI specifications.
//
// Summary: Discovery provider that uses OpenAPI specifications.
type OpenAPIProvider struct {
	Endpoint string // e.g., "http://localhost:8080/openapi.json"
}

// Name returns the name of the provider.
//
// Summary: Returns the canonical name for this OpenAPI discovery provider.
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
func (p *OpenAPIProvider) Name() string {
	return "openapi"
}

// Discover attempts to find services and return their configurations.
//
// Summary: Generates a dynamic service configuration for an OpenAPI upstream based on the provided specification endpoint.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The discovered service configurations.
//   - error: An error if discovery fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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
