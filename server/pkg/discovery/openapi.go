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
// Summary: OpenAPIProvider discovers services via OpenAPI specifications.
type OpenAPIProvider struct {
	Endpoint string // e.g., "http://localhost:8080/openapi.json"
}

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
func (p *OpenAPIProvider) Name() string {
	return "openapi"
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
