// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/catalog"
)

// CatalogServer represents the public CatalogServer entity.
//
// Summary: Provides network listening and request routing capabilities for server.
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
type CatalogServer struct {
	manager *catalog.Manager
}

// NewCatalogServer serves as a public interface for interacting with NewCatalogServer.
//
// Summary: Constructs and returns an initialized catalog server ready for consumption.
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
func NewCatalogServer(manager *catalog.Manager) *CatalogServer {
	return &CatalogServer{manager: manager}
}

// ListServices serves as a public interface for interacting with ListServices.
//
// Summary: List the services appropriately based on current system conditions.
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
func (s *CatalogServer) ListServices(ctx context.Context, _ *apiv1.ListCatalogServicesRequest) (*apiv1.ListCatalogServicesResponse, error) {
	services, err := s.manager.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	return apiv1.ListCatalogServicesResponse_builder{Services: services}.Build(), nil
}
