// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCatalogServer_ListServices verifies that the CatalogServer correctly lists services
// loaded by the catalog.Manager.
func TestCatalogServer_ListServices(t *testing.T) {
	// 1. Setup Virtual Filesystem
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// 2. Create Dummy Service Config
	configContent := `
upstream_services:
  - name: test-service-1
    http_service:
      address: http://example.com
`
	err = afero.WriteFile(fs, catalogPath+"/service1.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)
	err = manager.Load(context.Background())
	require.NoError(t, err)

	// 4. Initialize Catalog Server
	server := NewCatalogServer(manager)

	// 5. Test ListServices
	resp, err := server.ListServices(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Services, 1)

	assert.Equal(t, "test-service-1", resp.Services[0].GetName())
}

func TestCatalogServer_ListServices_Empty(t *testing.T) {
	// 1. Setup Virtual Filesystem
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// 2. Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)
	err = manager.Load(context.Background())
	require.NoError(t, err)

	// 3. Initialize Catalog Server
	server := NewCatalogServer(manager)

	// 4. Test ListServices
	resp, err := server.ListServices(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Services)
}

func TestCatalogServer_ListServices_InvalidConfig(t *testing.T) {
	// 1. Setup Virtual Filesystem
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// 2. Create Invalid Service Config (Malformed YAML)
	configContent := `
upstream_services:
  - name: test-service-1
    http_service:
      address: [ unclosed list
`
	err = afero.WriteFile(fs, catalogPath+"/invalid.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Initialize Catalog Manager
	manager := catalog.NewManager(fs, catalogPath)
	// Load should typically not fail entirely on one bad file, it might just log error and skip.
	// Based on manager.go code: "Log error but continue"
	err = manager.Load(context.Background())
	require.NoError(t, err) // Manager handles individual file errors gracefully

	// 4. Initialize Catalog Server
	server := NewCatalogServer(manager)

	// 5. Test ListServices - Should be empty as the only file was invalid
	resp, err := server.ListServices(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Services)
}
