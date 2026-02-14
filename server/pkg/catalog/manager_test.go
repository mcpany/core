// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Load_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a valid service config
	err := afero.WriteFile(fs, "catalog/service-a/config.yaml", []byte(`
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://localhost:8080"
`), 0644)
	require.NoError(t, err)

	// Create another valid service config
	err = afero.WriteFile(fs, "catalog/service-b/config.yaml", []byte(`
upstream_services:
  - name: "service-b"
    grpc_service:
      address: "localhost:9090"
`), 0644)
	require.NoError(t, err)

	m := NewManager(fs, "catalog")
	err = m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, 2)

	// Verify service-a
	var serviceAFound, serviceBFound bool
	for _, s := range services {
		if s.GetName() == "service-a" {
			serviceAFound = true
			assert.Equal(t, "http://localhost:8080", s.GetHttpService().GetAddress())
		} else if s.GetName() == "service-b" {
			serviceBFound = true
			assert.Equal(t, "localhost:9090", s.GetGrpcService().GetAddress())
		}
	}
	assert.True(t, serviceAFound, "service-a should be found")
	assert.True(t, serviceBFound, "service-b should be found")
}

func TestManager_Load_EmptyDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("catalog", 0755)
	require.NoError(t, err)

	m := NewManager(fs, "catalog")
	err = m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Empty(t, services)
}

func TestManager_Load_IgnoreNonYaml(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a non-YAML file
	err := afero.WriteFile(fs, "catalog/README.md", []byte("# Catalog"), 0644)
	require.NoError(t, err)

	// Create a valid YAML file
	err = afero.WriteFile(fs, "catalog/service-a/config.yaml", []byte(`
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://localhost:8080"
`), 0644)
	require.NoError(t, err)

	m := NewManager(fs, "catalog")
	err = m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "service-a", services[0].GetName())
}

func TestManager_Load_InvalidConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create an invalid YAML file (broken syntax)
	err := afero.WriteFile(fs, "catalog/broken/config.yaml", []byte(`
upstream_services:
  - name: "broken"
    http_service:
      address: "http://localhost:8080"
    broken_field:
`), 0644)
	require.NoError(t, err)

	// Create a valid YAML file
	err = afero.WriteFile(fs, "catalog/valid/config.yaml", []byte(`
upstream_services:
  - name: "valid"
    http_service:
      address: "http://localhost:8080"
`), 0644)
	require.NoError(t, err)

	m := NewManager(fs, "catalog")
	// Load should succeed even if individual files fail
	err = m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	// Only the valid service should be loaded
	assert.Len(t, services, 1)
	assert.Equal(t, "valid", services[0].GetName())
}

func TestManager_ListServices_Concurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "catalog/service-a/config.yaml", []byte(`
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://localhost:8080"
`), 0644)
	require.NoError(t, err)

	m := NewManager(fs, "catalog")

	// Start a goroutine that repeatedly loads the catalog
	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = m.Load(context.Background())
				// Sleep a tiny bit to yield
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Read concurrently
	for i := 0; i < 100; i++ {
		services, err := m.ListServices(context.Background())
		assert.NoError(t, err)
		// It should be either empty (before first load) or have 1 service
		if len(services) > 0 {
			assert.Equal(t, "service-a", services[0].GetName())
		}
		time.Sleep(1 * time.Millisecond)
	}

	close(stop)
	wg.Wait()
}
