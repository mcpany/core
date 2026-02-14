// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	manager := NewManager(fs, "/catalog")
	assert.NotNil(t, manager)
}

func TestManager_Load_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	catalogPath := "/catalog"

	// Create a valid service config
	validConfig := `
upstream_services:
  - name: "service-1"
    http_service:
      address: "http://localhost:8080"
`
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/service1.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, catalogPath)
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, "service-1", services[0].GetName())
	assert.Equal(t, "http://localhost:8080", services[0].GetHttpService().GetAddress())
}

func TestManager_Load_MultipleFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	catalogPath := "/catalog"

	config1 := `
upstream_services:
  - name: "service-1"
    http_service:
      address: "http://localhost:8081"
`
	config2 := `
upstream_services:
  - name: "service-2"
    grpc_service:
      address: "localhost:9090"
`

	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/service1.yaml", []byte(config1), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/service2.yml", []byte(config2), 0644) // .yml extension
	require.NoError(t, err)

	manager := NewManager(fs, catalogPath)
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	require.Len(t, services, 2)

	// Verify both services are present (order depends on filesystem walk, but usually sorted or consistent in mock)
	names := make(map[string]bool)
	for _, s := range services {
		names[s.GetName()] = true
	}
	assert.True(t, names["service-1"])
	assert.True(t, names["service-2"])
}

func TestManager_Load_EmptyDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	catalogPath := "/catalog"

	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	manager := NewManager(fs, catalogPath)
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Empty(t, services)
}

func TestManager_Load_InvalidYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	catalogPath := "/catalog"

	// Invalid YAML (tab character is invalid in YAML, or just bad structure)
	invalidConfig := `
upstream_services:
  - name: "service-1"
    http_service:
	  address: "http://localhost:8080" # Tab indentation
`
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/invalid.yaml", []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Also add a valid one to ensure it continues
	validConfig := `
upstream_services:
  - name: "service-2"
    http_service:
      address: "http://localhost:8082"
`
	err = afero.WriteFile(fs, catalogPath+"/valid.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, catalogPath)
	err = manager.Load(ctx)
	require.NoError(t, err) // Should not return error, just log it

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, "service-2", services[0].GetName())
}

func TestManager_Load_NonYAMLFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	catalogPath := "/catalog"

	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/README.md", []byte("# Catalog"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/script.sh", []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	manager := NewManager(fs, catalogPath)
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Empty(t, services)
}

func TestManager_ListServices_Copy(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	catalogPath := "/catalog"

	validConfig := `
upstream_services:
  - name: "service-1"
    http_service:
      address: "http://localhost:8080"
`
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, catalogPath+"/service1.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, catalogPath)
	err = manager.Load(ctx)
	require.NoError(t, err)

	services1, _ := manager.ListServices(ctx)
	services2, _ := manager.ListServices(ctx)

	assert.Equal(t, services1, services2)
	assert.NotSame(t, &services1, &services2) // Slices are different headers
	// Note: The pointers inside the slice are the same because ListServices creates a shallow copy of the slice.
	// If we modify the underlying struct, it would reflect in both.
	// But adding/removing from the slice returned by ListServices won't affect the manager's internal slice.

	services1[0] = &configv1.UpstreamServiceConfig{} // Replace element in returned slice

	services3, _ := manager.ListServices(ctx)
	assert.NotEqual(t, services1[0], services3[0]) // Manager's state should be preserved
}

func TestManager_Load_WalkError(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	// Do not create the directory
	manager := NewManager(fs, "/non-existent")
	err := manager.Load(ctx)
	assert.Error(t, err)
}
