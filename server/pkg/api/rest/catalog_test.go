// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"path/filepath"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogServer_ListServices(t *testing.T) {
	// 1. Create in-memory filesystem
	fs := afero.NewMemMapFs()

	// 2. Define catalog path
	catalogPath := "/catalog"

	// 3. Create a valid service configuration
	serviceConfig := `
upstream_services:
  - id: "test-service-1"
    name: "Test Service 1"
    version: "1.0.0"
`
	// Write config file
	configPath := filepath.Join(catalogPath, "service1", "config.yaml")
	err := fs.MkdirAll(filepath.Dir(configPath), 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, configPath, []byte(serviceConfig), 0644)
	require.NoError(t, err)

	// 4. Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)

	// 5. Load services
	ctx := context.Background()
	err = manager.Load(ctx)
	require.NoError(t, err)

	// 6. Initialize Catalog Server
	server := NewCatalogServer(manager)

	// 7. Call ListServices
	req := &apiv1.ListCatalogServicesRequest{}
	resp, err := server.ListServices(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// 8. Verify Results
	assert.Len(t, resp.Services, 1)
	svc := resp.Services[0]
	assert.Equal(t, "test-service-1", svc.GetId())
	assert.Equal(t, "Test Service 1", svc.GetName())
	assert.Equal(t, "1.0.0", svc.GetVersion())
}

func TestCatalogServer_ListServices_Empty(t *testing.T) {
	// 1. Create in-memory filesystem (empty)
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// 2. Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)

	// 3. Load services (empty)
	ctx := context.Background()
	err = manager.Load(ctx)
	require.NoError(t, err)

	// 4. Initialize Catalog Server
	server := NewCatalogServer(manager)

	// 5. Call ListServices
	req := &apiv1.ListCatalogServicesRequest{}
	resp, err := server.ListServices(ctx, req)
	require.NoError(t, err)

	// 6. Verify Results
	assert.Empty(t, resp.Services)
}
