// Copyright 2026 Author(s) of MCP Any
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
	// Setup FS
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// Write a sample service config
	// Note: The YAML parser expects snake_case keys which match proto field names or json_name
	svcConfig := `
upstream_services:
  - id: "svc-1"
    name: "service1"
    version: "1.0.0"
`
	err = afero.WriteFile(fs, catalogPath+"/service1.yaml", []byte(svcConfig), 0644)
	require.NoError(t, err)

	// Initialize Manager
	manager := catalog.NewManager(fs, catalogPath)
	err = manager.Load(context.Background())
	require.NoError(t, err)

	// Initialize Server
	server := NewCatalogServer(manager)

	// Call ListServices
	resp, err := server.ListServices(context.Background(), &apiv1.ListCatalogServicesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Services, 1)

	svc := resp.Services[0]
	assert.Equal(t, "svc-1", svc.GetId())
	assert.Equal(t, "service1", svc.GetName())
	assert.Equal(t, "1.0.0", svc.GetVersion())
}

func TestCatalogServer_ListServices_Empty(t *testing.T) {
	// Setup FS (Empty)
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// Initialize Manager
	manager := catalog.NewManager(fs, catalogPath)
	err = manager.Load(context.Background())
	require.NoError(t, err)

	// Initialize Server
	server := NewCatalogServer(manager)

	// Call ListServices
	resp, err := server.ListServices(context.Background(), &apiv1.ListCatalogServicesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Services)
}
