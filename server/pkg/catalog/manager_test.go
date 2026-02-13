// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"

	m := NewManager(fs, catalogPath)

	assert.NotNil(t, m)
	assert.Equal(t, fs, m.fs)
	assert.Equal(t, catalogPath, m.catalogPath)
}

func TestManager_Load(t *testing.T) {
	ctx := context.Background()

	t.Run("Empty Directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		catalogPath := "/catalog"
		err := fs.MkdirAll(catalogPath, 0755)
		require.NoError(t, err)

		m := NewManager(fs, catalogPath)
		err = m.Load(ctx)
		require.NoError(t, err)

		services, err := m.ListServices(ctx)
		require.NoError(t, err)
		assert.Empty(t, services)
	})

	t.Run("Valid Configs", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		catalogPath := "/catalog"
		err := fs.MkdirAll(catalogPath, 0755)
		require.NoError(t, err)

		// Create service1.yaml
		service1 := `
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`
		err = afero.WriteFile(fs, filepath.Join(catalogPath, "service1.yaml"), []byte(service1), 0644)
		require.NoError(t, err)

		// Create service2.yml
		service2 := `
upstream_services:
  - name: "service2"
    grpc_service:
      address: "localhost:50051"
`
		err = afero.WriteFile(fs, filepath.Join(catalogPath, "service2.yml"), []byte(service2), 0644)
		require.NoError(t, err)

		m := NewManager(fs, catalogPath)
		err = m.Load(ctx)
		require.NoError(t, err)

		services, err := m.ListServices(ctx)
		require.NoError(t, err)
		assert.Len(t, services, 2)

		names := make([]string, 0, len(services))
		for _, s := range services {
			names = append(names, s.GetName())
		}
		assert.Contains(t, names, "service1")
		assert.Contains(t, names, "service2")
	})

	t.Run("Invalid Config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		catalogPath := "/catalog"
		err := fs.MkdirAll(catalogPath, 0755)
		require.NoError(t, err)

		// Valid service
		service1 := `
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`
		err = afero.WriteFile(fs, filepath.Join(catalogPath, "service1.yaml"), []byte(service1), 0644)
		require.NoError(t, err)

		// Invalid service (bad yaml)
		badService := `
upstream_services:
  - name: "bad"
    http_service:
      address: "http://bad.com"
  INVALID_INDENTATION
`
		err = afero.WriteFile(fs, filepath.Join(catalogPath, "bad.yaml"), []byte(badService), 0644)
		require.NoError(t, err)

		m := NewManager(fs, catalogPath)
		err = m.Load(ctx)
		require.NoError(t, err) // Should not fail overall

		services, err := m.ListServices(ctx)
		require.NoError(t, err)
		assert.Len(t, services, 1) // Only service1 should be loaded
		assert.Equal(t, "service1", services[0].GetName())
	})

	t.Run("Ignore Non-YAML", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		catalogPath := "/catalog"
		err := fs.MkdirAll(catalogPath, 0755)
		require.NoError(t, err)

		readme := "This is a readme file."
		err = afero.WriteFile(fs, filepath.Join(catalogPath, "README.md"), []byte(readme), 0644)
		require.NoError(t, err)

		m := NewManager(fs, catalogPath)
		err = m.Load(ctx)
		require.NoError(t, err)

		services, err := m.ListServices(ctx)
		require.NoError(t, err)
		assert.Empty(t, services)
	})

	t.Run("Recurse Subdirectories", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		catalogPath := "/catalog"
		subDir := filepath.Join(catalogPath, "subdir")
		err := fs.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		// Config in subdirectory
		service1 := `
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`
		err = afero.WriteFile(fs, filepath.Join(subDir, "service1.yaml"), []byte(service1), 0644)
		require.NoError(t, err)

		m := NewManager(fs, catalogPath)
		err = m.Load(ctx)
		require.NoError(t, err)

		services, err := m.ListServices(ctx)
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "service1", services[0].GetName())
	})
}

func TestManager_Concurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	service1 := `
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`
	err = afero.WriteFile(fs, filepath.Join(catalogPath, "service1.yaml"), []byte(service1), 0644)
	require.NoError(t, err)

	m := NewManager(fs, catalogPath)
	ctx := context.Background()

	// Initial Load
	err = m.Load(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	// Run parallel Load and ListServices
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := m.ListServices(ctx)
			assert.NoError(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.Load(ctx)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}
