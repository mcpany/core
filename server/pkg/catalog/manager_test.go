// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Load(t *testing.T) {
	fs := afero.NewMemMapFs()
	catalogPath := "/catalog"

	// Create catalog directory
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// Case 1: Empty Directory
	t.Run("Empty Directory", func(t *testing.T) {
		manager := NewManager(fs, catalogPath)
		err := manager.Load(context.Background())
		require.NoError(t, err)
		services, err := manager.ListServices(context.Background())
		require.NoError(t, err)
		assert.Empty(t, services)
	})

	// Case 2: Happy Path - Valid YAML
	t.Run("Valid YAML", func(t *testing.T) {
		validYaml := `
upstream_services:
  - name: "service-1"
    http_service:
      address: "http://example.com"
`
		err := afero.WriteFile(fs, filepath.Join(catalogPath, "service1.yaml"), []byte(validYaml), 0644)
		require.NoError(t, err)

		manager := NewManager(fs, catalogPath)
		err = manager.Load(context.Background())
		require.NoError(t, err)
		services, err := manager.ListServices(context.Background())
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "service-1", services[0].GetName())
		assert.Equal(t, "http://example.com", services[0].GetHttpService().GetAddress())
	})

	// Case 3: Invalid YAML - Should Log Error but Continue
	t.Run("Invalid YAML", func(t *testing.T) {
		invalidYaml := `
upstream_services:
  - name: "service-broken"
    http_service:
      address: [this is broken yaml
`
		err := afero.WriteFile(fs, filepath.Join(catalogPath, "broken.yaml"), []byte(invalidYaml), 0644)
		require.NoError(t, err)

		// Add another valid one to ensure it continues
		validYaml2 := `
upstream_services:
  - name: "service-2"
    http_service:
      address: "http://example.org"
`
		err = afero.WriteFile(fs, filepath.Join(catalogPath, "service2.yaml"), []byte(validYaml2), 0644)
		require.NoError(t, err)

		manager := NewManager(fs, catalogPath)
		err = manager.Load(context.Background())
		require.NoError(t, err) // Should not fail overall

		services, err := manager.ListServices(context.Background())
		require.NoError(t, err)

		// Total services: 2 (service-1 from previous test + service-2)
		assert.Len(t, services, 2)
		names := []string{services[0].GetName(), services[1].GetName()}
		assert.Contains(t, names, "service-1")
		assert.Contains(t, names, "service-2")
	})

	// Case 4: Non-YAML Files - Should Ignore
	t.Run("Non-YAML Files", func(t *testing.T) {
		err := afero.WriteFile(fs, filepath.Join(catalogPath, "readme.txt"), []byte("Just a text file"), 0644)
		require.NoError(t, err)

		manager := NewManager(fs, catalogPath)
		err = manager.Load(context.Background())
		require.NoError(t, err)

		services, err := manager.ListServices(context.Background())
		require.NoError(t, err)
		// Still 2 valid services
		assert.Len(t, services, 2)
	})

	// Case 5: Nested Directories
	t.Run("Nested Directories", func(t *testing.T) {
		nestedPath := filepath.Join(catalogPath, "nested")
		err := fs.MkdirAll(nestedPath, 0755)
		require.NoError(t, err)

		validYaml3 := `
upstream_services:
  - name: "service-3"
    http_service:
      address: "http://nested.example.com"
`
		err = afero.WriteFile(fs, filepath.Join(nestedPath, "service3.yml"), []byte(validYaml3), 0644)
		require.NoError(t, err)

		manager := NewManager(fs, catalogPath)
		err = manager.Load(context.Background())
		require.NoError(t, err)

		services, err := manager.ListServices(context.Background())
		require.NoError(t, err)

		// Total: service-1, service-2, service-3
		assert.Len(t, services, 3)
		names := []string{services[0].GetName(), services[1].GetName(), services[2].GetName()}
		assert.Contains(t, names, "service-3")
	})
}
