// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	catalogPath := "marketplace/catalog"
	manager := NewManager(fs, catalogPath)

	assert.NotNil(t, manager)
	assert.Equal(t, fs, manager.fs)
	assert.Equal(t, catalogPath, manager.catalogPath)
	assert.Nil(t, manager.services)
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string]string
		expectError   bool
		expectService int
		verify        func(t *testing.T, manager *Manager)
	}{
		{
			name:        "Empty Directory",
			files:       map[string]string{},
			expectError: false,
			verify: func(t *testing.T, manager *Manager) {
				services, err := manager.ListServices(context.Background())
				assert.NoError(t, err)
				assert.Empty(t, services)
			},
		},
		{
			name: "Valid Config",
			files: map[string]string{
				"marketplace/catalog/service1.yaml": `
upstream_services:
  - name: service1
    id: service1
    http_service:
      address: http://localhost:8080
`,
			},
			expectError: false,
			verify: func(t *testing.T, manager *Manager) {
				services, err := manager.ListServices(context.Background())
				assert.NoError(t, err)
				assert.Len(t, services, 1)
				assert.Equal(t, "service1", services[0].GetName())
			},
		},
		{
			name: "Invalid YAML",
			files: map[string]string{
				"marketplace/catalog/service1.yaml": `
upstream_services:
  - name: service1
    id: service1
    http_service:
      address: http://localhost:8080
`,
				"marketplace/catalog/invalid.yaml": `
invalid_yaml: [
`,
			},
			expectError: false, // Load logs error but continues
			verify: func(t *testing.T, manager *Manager) {
				services, err := manager.ListServices(context.Background())
				assert.NoError(t, err)
				assert.Len(t, services, 1)
				assert.Equal(t, "service1", services[0].GetName())
			},
		},
		{
			name: "No Services in Config",
			files: map[string]string{
				"marketplace/catalog/empty.yaml": `
global_settings:
  log_level: DEBUG
`,
			},
			expectError: false,
			verify: func(t *testing.T, manager *Manager) {
				services, err := manager.ListServices(context.Background())
				assert.NoError(t, err)
				assert.Empty(t, services)
			},
		},
		{
			name: "Multiple Files",
			files: map[string]string{
				"marketplace/catalog/service1.yaml": `
upstream_services:
  - name: service1
    id: service1
`,
				"marketplace/catalog/service2.yaml": `
upstream_services:
  - name: service2
    id: service2
`,
			},
			expectError: false,
			verify: func(t *testing.T, manager *Manager) {
				services, err := manager.ListServices(context.Background())
				assert.NoError(t, err)
				assert.Len(t, services, 2)
				names := []string{services[0].GetName(), services[1].GetName()}
				assert.Contains(t, names, "service1")
				assert.Contains(t, names, "service2")
			},
		},
		{
			name: "Ignored Files",
			files: map[string]string{
				"marketplace/catalog/service1.txt": `
upstream_services:
  - name: service1
`,
			},
			expectError: false,
			verify: func(t *testing.T, manager *Manager) {
				services, err := manager.ListServices(context.Background())
				assert.NoError(t, err)
				assert.Empty(t, services)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			catalogPath := "marketplace/catalog"
			manager := NewManager(fs, catalogPath)

			for path, content := range tt.files {
				dir := filepath.Dir(path)
				err := fs.MkdirAll(dir, 0755)
				require.NoError(t, err)
				err = afero.WriteFile(fs, path, []byte(content), 0644)
				require.NoError(t, err)
			}

			// Ensure directory exists even if empty
			if len(tt.files) == 0 {
				err := fs.MkdirAll(catalogPath, 0755)
				require.NoError(t, err)
			}

			err := manager.Load(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.verify != nil {
				tt.verify(t, manager)
			}
		})
	}
}

func TestListServices_Copy(t *testing.T) {
	fs := afero.NewMemMapFs()
	catalogPath := "marketplace/catalog"
	manager := NewManager(fs, catalogPath)

	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "marketplace/catalog/service1.yaml", []byte(`
upstream_services:
  - name: service1
`), 0644)
	require.NoError(t, err)

	err = manager.Load(context.Background())
	require.NoError(t, err)

	services1, _ := manager.ListServices(context.Background())
	services2, _ := manager.ListServices(context.Background())

	// Verify slices contain pointers to the same objects (shallow copy)
	assert.Equal(t, services1[0], services2[0])

	// Verify backing arrays are different by mutating one slice's element
	// If they shared the backing array, setting services1[0] to nil would affect services2[0]
	services1[0] = nil
	assert.NotNil(t, services2[0], "services2 should not be affected by mutation of services1")
}

func TestConcurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	catalogPath := "marketplace/catalog"
	manager := NewManager(fs, catalogPath)
	err := fs.MkdirAll(catalogPath, 0755)
	require.NoError(t, err)

	// Write initial file
	err = afero.WriteFile(fs, "marketplace/catalog/service1.yaml", []byte(`
upstream_services:
  - name: service1
`), 0644)
	require.NoError(t, err)

	err = manager.Load(context.Background())
	require.NoError(t, err)

	var wg sync.WaitGroup
	ctx := context.Background()

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				services, err := manager.ListServices(ctx)
				assert.NoError(t, err)
				assert.NotEmpty(t, services)
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Writers (Reloads)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				// Update file content slightly to simulate change
				_ = afero.WriteFile(fs, "marketplace/catalog/service1.yaml", []byte(`
upstream_services:
  - name: service1_updated
`), 0644)
				err := manager.Load(ctx)
				assert.NoError(t, err)
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
}
