// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/catalog"
)

// CatalogServer implements the CatalogService API.
type CatalogServer struct {
	manager *catalog.Manager
}

// NewCatalogServer creates a new CatalogServer.
func NewCatalogServer(manager *catalog.Manager) *CatalogServer {
	return &CatalogServer{manager: manager}
}

// ListServices returns a list of available services in the catalog.
func (s *CatalogServer) ListServices(ctx context.Context, _ *apiv1.ListCatalogServicesRequest) (*apiv1.ListCatalogServicesResponse, error) {
	services, err := s.manager.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	return &apiv1.ListCatalogServicesResponse{Services: services}, nil
}
