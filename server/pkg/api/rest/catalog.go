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
// Summary: Server implementation for the Catalog Service.
//
// It handles requests to list available services from the dynamic catalog.
type CatalogServer struct {
	manager *catalog.Manager
}

// NewCatalogServer provides newcatalogserver functionality.
//
// Summary: NewCatalogServer.
//
// Parameters.
//   - manager: The parameter.
//
// Returns.
//   - result: The result.
func NewCatalogServer(manager *catalog.Manager) *CatalogServer {
	return &CatalogServer{manager: manager}
}

// ListServices provides listservices functionality.
//
// Summary: ListServices.
//
// Parameters.
//   - ctx: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *CatalogServer) ListServices(ctx context.Context, _ *apiv1.ListCatalogServicesRequest) (*apiv1.ListCatalogServicesResponse, error) {
	services, err := s.manager.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	return apiv1.ListCatalogServicesResponse_builder{Services: services}.Build(), nil
}
