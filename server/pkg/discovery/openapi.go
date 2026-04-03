// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// OpenAPIProvider represents the public OpenAPIProvider entity.
//
// Summary: Defines the structured data model representing a api provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type OpenAPIProvider struct {
	Endpoint string // e.g., "http://localhost:8080/openapi.json"
}

// Name serves as a public interface for interacting with Name.
//
// Summary: Name the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *OpenAPIProvider) Name() string {
	return "openapi"
}

// Discover serves as a public interface for interacting with Discover.
//
// Summary: Discover the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
