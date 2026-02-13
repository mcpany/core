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

func TestNewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/catalog"
	manager := NewManager(fs, path)

	assert.NotNil(t, manager)
	assert.Equal(t, fs, manager.fs)
	assert.Equal(t, path, manager.catalogPath)
}

func TestManager_Load(t *testing.T) {
	tests := []struct {
		name           string
		setupFs        func(fs afero.Fs)
		expectedCount  int
		expectedNames  []string
		expectedErrors bool
	}{
		{
			name:    "Empty Directory",
			setupFs: func(fs afero.Fs) {}, // Do nothing
			expectedCount: 0,
		},
		{
			name: "Valid Configs",
			setupFs: func(fs afero.Fs) {
				// Create service1.yaml
				afero.WriteFile(fs, "/catalog/service1.yaml", []byte(`
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`), 0644)
				// Create service2.yml
				afero.WriteFile(fs, "/catalog/service2.yml", []byte(`
upstream_services:
  - name: "service2"
    grpc_service:
      address: "localhost:50051"
`), 0644)
			},
			expectedCount: 2,
			expectedNames: []string{"service1", "service2"},
		},
		{
			name: "Recursive Loading",
			setupFs: func(fs afero.Fs) {
				fs.MkdirAll("/catalog/subdir", 0755)
				afero.WriteFile(fs, "/catalog/subdir/nested.yaml", []byte(`
upstream_services:
  - name: "nested"
    http_service:
      address: "http://nested.com"
`), 0644)
			},
			expectedCount: 1,
			expectedNames: []string{"nested"},
		},
		{
			name: "Invalid Config",
			setupFs: func(fs afero.Fs) {
				// Valid file
				afero.WriteFile(fs, "/catalog/valid.yaml", []byte(`
upstream_services:
  - name: "valid"
    http_service:
      address: "http://valid.com"
`), 0644)
				// Invalid YAML
				afero.WriteFile(fs, "/catalog/bad.yaml", []byte(`
upstream_services:
  - name: "bad"
    http_service:
      address: "http://bad.com"
  INVALID_YAML_SYNTAX
`), 0644)
			},
			expectedCount: 1, // Only the valid one should load
			expectedNames: []string{"valid"},
		},
		{
			name: "Ignore Non-YAML",
			setupFs: func(fs afero.Fs) {
				afero.WriteFile(fs, "/catalog/readme.txt", []byte("This is not a config"), 0644)
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			fs.MkdirAll("/catalog", 0755)
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			m := NewManager(fs, "/catalog")
			err := m.Load(context.Background())

			if tt.expectedErrors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			services, err := m.ListServices(context.Background())
			assert.NoError(t, err)
			assert.Len(t, services, tt.expectedCount)

			if tt.expectedNames != nil {
				var names []string
				for _, s := range services {
					names = append(names, s.GetName())
				}
				assert.ElementsMatch(t, tt.expectedNames, names)
			}
		})
	}
}

func TestManager_Concurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/catalog", 0755)
	afero.WriteFile(fs, "/catalog/service1.yaml", []byte(`
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`), 0644)

	m := NewManager(fs, "/catalog")
	ctx := context.Background()

	// Initial load
	require.NoError(t, m.Load(ctx))

	var wg sync.WaitGroup
	// Simulate concurrent reads and writes (Load is a write, ListServices is a read)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Randomly sleep to create race conditions
			// But we want to test race, so just loop
			for j := 0; j < 10; j++ {
				_, err := m.ListServices(ctx)
				assert.NoError(t, err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				// Re-load periodically
				assert.NoError(t, m.Load(ctx))
			}
		}()
	}

	// Wait with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for concurrency test")
	}
}
