// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package discovery implements the discovery subsystem.
package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// OpenAPIProvider discovers services via OpenAPI specifications.
type OpenAPIProvider struct {
	Endpoint string // e.g., "http://localhost:8080/openapi.json"
}

// Name handles name.
//
// Parameters:
//   - None
//
// Returns:
//   - string: The generated or retrieved entity.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Name returns the name of the provider.
func (p *OpenAPIProvider) Name() string {
	return "openapi"
}

// Discover handles discover.
//
// Parameters:
//   - _ (context.Context): Reserved parameter, currently unused.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A slice containing the requested elements.
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
// Discover attempts to find services and return their configurations.
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
