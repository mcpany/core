// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/catalog"
)

// CatalogServer implements the CatalogService API.
//
// Summary: Handles catalog-related API requests.
type CatalogServer struct {
	manager *catalog.Manager
}

// NewCatalogServer creates a new CatalogServer.
//
// Summary: Initializes a new CatalogServer.
//
// Parameters:
//   - manager: *catalog.Manager. The catalog manager instance.
//
// Returns:
//   - *CatalogServer: The initialized server.
func NewCatalogServer(manager *catalog.Manager) *CatalogServer {
	return &CatalogServer{manager: manager}
}

// ListServices returns a list of available services in the catalog.
//
// Summary: Lists all services in the catalog.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - _: *apiv1.ListCatalogServicesRequest. The request object (unused).
//
// Returns:
//   - *apiv1.ListCatalogServicesResponse: The response containing the list of services.
//   - error: An error if the operation fails.
func (s *CatalogServer) ListServices(ctx context.Context, _ *apiv1.ListCatalogServicesRequest) (*apiv1.ListCatalogServicesResponse, error) {
	services, err := s.manager.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	return &apiv1.ListCatalogServicesResponse{Services: services}, nil
}
