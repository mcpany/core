// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogServer_ListServices(t *testing.T) {
	// Setup Mock FS
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	require.NoError(t, fs.MkdirAll(catalogPath, 0755))

	// Create a dummy service config
	serviceConfig := `
upstream_services:
  - name: "test-service"
    http_service:
      address: "http://example.com"
`
	require.NoError(t, afero.WriteFile(fs, catalogPath+"/service.yaml", []byte(serviceConfig), 0644))

	// Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)
	ctx := context.Background()
	require.NoError(t, manager.Load(ctx))

	// Initialize Server
	server := NewCatalogServer(manager)

	// Test ListServices
	req := &apiv1.ListCatalogServicesRequest{}
	resp, err := server.ListServices(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify
	assert.Len(t, resp.Services, 1)
	assert.Equal(t, "test-service", resp.Services[0].GetName())
}

func TestCatalogServer_ListServices_Empty(t *testing.T) {
	// Setup Mock FS (Empty)
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	require.NoError(t, fs.MkdirAll(catalogPath, 0755))

	// Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)
	ctx := context.Background()
	require.NoError(t, manager.Load(ctx))

	// Initialize Server
	server := NewCatalogServer(manager)

	// Test ListServices
	req := &apiv1.ListCatalogServicesRequest{}
	resp, err := server.ListServices(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Services)
}
