// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	m := NewManager(fs, "/config")
	assert.NotNil(t, m)
	// access private field via reflection or just trust it's set if public methods work
	// Since catalogPath is private, we can't assert it directly without reflection or changing visibility.
	// But NewManager is simple enough.
}

func TestManager_Load(t *testing.T) {
	tests := []struct {
		name         string
		setupFs      func(fs afero.Fs)
		wantServices int
		wantErr      bool
	}{
		{
			name:         "Empty Directory",
			setupFs:      func(fs afero.Fs) { _ = fs.MkdirAll("/config", 0755) },
			wantServices: 0,
			wantErr:      false,
		},
		{
			name: "Valid Configs",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/config", 0755)
				_ = afero.WriteFile(fs, "/config/service1.yaml", []byte(`
upstream_services:
  - name: "service1"
    http_service:
      address: "http://service1.com"
`), 0644)
				_ = afero.WriteFile(fs, "/config/service2.yml", []byte(`
upstream_services:
  - name: "service2"
    http_service:
      address: "http://service2.com"
`), 0644)
			},
			wantServices: 2,
			wantErr:      false,
		},
		{
			name: "Invalid Config",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/config", 0755)
				_ = afero.WriteFile(fs, "/config/bad.yaml", []byte(`invalid_yaml: [`), 0644)
				// Also add a valid one to make sure it still loads valid ones
				_ = afero.WriteFile(fs, "/config/service1.yaml", []byte(`
upstream_services:
  - name: "service1"
    http_service:
      address: "http://service1.com"
`), 0644)
			},
			wantServices: 1, // The valid one should be loaded
			wantErr:      false,
		},
		{
			name: "Ignore Non-YAML",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/config", 0755)
				_ = afero.WriteFile(fs, "/config/readme.txt", []byte(`some text`), 0644)
			},
			wantServices: 0,
			wantErr:      false,
		},
		{
			name: "Recurse Subdirectories",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/config/subdir", 0755)
				_ = afero.WriteFile(fs, "/config/subdir/service3.yaml", []byte(`
upstream_services:
  - name: "service3"
    http_service:
      address: "http://service3.com"
`), 0644)
			},
			wantServices: 1,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setupFs(fs)
			m := NewManager(fs, "/config")

			err := m.Load(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			services, err := m.ListServices(context.Background())
			assert.NoError(t, err)
			assert.Len(t, services, tt.wantServices)

			// Additional check for service content
			if tt.wantServices > 0 {
				for _, svc := range services {
					assert.IsType(t, &configv1.UpstreamServiceConfig{}, svc)
					assert.NotEmpty(t, svc.GetName())
				}
			}
		})
	}
}

func TestManager_Concurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/config", 0755)
	_ = afero.WriteFile(fs, "/config/service1.yaml", []byte(`
upstream_services:
  - name: "service1"
    http_service:
      address: "http://service1.com"
`), 0644)

	m := NewManager(fs, "/config")
	ctx := context.Background()

	// Initial load
	err := m.Load(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	// Simulate concurrent reads and writes (Load is write, ListServices is read)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Randomly call Load or ListServices
			if i%10 == 0 {
				if err := m.Load(ctx); err != nil {
					t.Errorf("Load failed: %v", err)
				}
			} else {
				if _, err := m.ListServices(ctx); err != nil {
					t.Errorf("ListServices failed: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()
}
