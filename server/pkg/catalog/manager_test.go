// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Load_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()

	// Create a valid service config
	validConfig := `
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://localhost:8080"
`
	err := afero.WriteFile(fs, "/catalog/service-a/config.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, "/catalog")
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "service-a", services[0].GetName())
	assert.Equal(t, "http://localhost:8080", services[0].GetHttpService().GetAddress())
}

func TestManager_Load_EmptyDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()

	// Create an empty directory
	err := fs.MkdirAll("/catalog", 0755)
	require.NoError(t, err)

	manager := NewManager(fs, "/catalog")
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Empty(t, services)
}

func TestManager_Load_IgnoreNonYaml(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()

	// Create a non-YAML file
	err := afero.WriteFile(fs, "/catalog/README.md", []byte("# Documentation"), 0644)
	require.NoError(t, err)

	// Create a valid YAML file
	validConfig := `
upstream_services:
  - name: "service-b"
    http_service:
      address: "http://localhost:8081"
`
	err = afero.WriteFile(fs, "/catalog/service-b/config.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, "/catalog")
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "service-b", services[0].GetName())
}

func TestManager_Load_InvalidConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()

	// Create a broken YAML file (unclosed string/broken syntax)
	brokenConfig := `
upstream_services:
  - name: "service-c"
    http_service:
      address: "http://localho
`
	err := afero.WriteFile(fs, "/catalog/service-c/config.yaml", []byte(brokenConfig), 0644)
	require.NoError(t, err)

	// Create a valid one too to ensure load continues
	validConfig := `
upstream_services:
  - name: "service-d"
    http_service:
      address: "http://localhost:8083"
`
	err = afero.WriteFile(fs, "/catalog/service-d/config.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, "/catalog")
	err = manager.Load(ctx)
	// The Manager.Load logs errors but continues. It returns error only on Walk failure.
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "service-d", services[0].GetName())
}

func TestManager_Load_NestedDirectories(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()

	// Nested structure: /catalog/category/service/config.yaml
	validConfig := `
upstream_services:
  - name: "service-nested"
    http_service:
      address: "http://nested:8080"
`
	err := afero.WriteFile(fs, "/catalog/category/service/config.yaml", []byte(validConfig), 0644)
	require.NoError(t, err)

	manager := NewManager(fs, "/catalog")
	err = manager.Load(ctx)
	require.NoError(t, err)

	services, err := manager.ListServices(ctx)
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "service-nested", services[0].GetName())
}

func TestManager_ListServices_Concurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	manager := NewManager(fs, "/catalog")

	// Start a goroutine that repeatedly calls ListServices
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_, _ = manager.ListServices(ctx)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Simulate repeated loads
	for i := 0; i < 10; i++ {
		validConfig := `
upstream_services:
  - name: "service-concurrency"
    http_service:
      address: "http://concurrency:8080"
`
		_ = afero.WriteFile(fs, "/catalog/service/config.yaml", []byte(validConfig), 0644)
		err := manager.Load(ctx)
		require.NoError(t, err)
		time.Sleep(5 * time.Millisecond)
	}
	close(done)
}

func TestManager_NewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	manager := NewManager(fs, "/test-path")
	assert.NotNil(t, manager)
	assert.Equal(t, "/test-path", manager.catalogPath)
}

func TestManager_Load_NonExistentDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := context.Background()

	// Do not create the directory
	manager := NewManager(fs, "/non-existent")
	err := manager.Load(ctx)
	// afero.Walk returns an error if root doesn't exist
	require.Error(t, err)
}
