// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"

// OpenAPIProvider discovers services via OpenAPI specifications.
//
// Summary: Represents a OpenAPIProvider.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type OpenAPIProvider struct {
	Endpoint string // e.g., "http://localhost:8080/openapi.json"
}

// Name returns the name of the provider.
//
// Summary: Executes Name operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (p *OpenAPIProvider) Name() string {
	return "openapi"
}

// Discover attempts to find services and return their configurations.
//
// Summary: Executes Discover operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
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
