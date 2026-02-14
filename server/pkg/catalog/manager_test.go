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

func TestManager_Load(t *testing.T) {
	tests := []struct {
		name        string
		setupFs     func(fs afero.Fs)
		verify      func(t *testing.T, services []*configv1.UpstreamServiceConfig)
		expectedErr bool
	}{
		{
			name: "Happy Path - Valid Configs",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/catalog", 0755))
				require.NoError(t, afero.WriteFile(fs, "/catalog/service1.yaml", []byte(`
upstream_services:
  - name: "service1"
    http_service:
      address: "http://service1"
`), 0644))
				require.NoError(t, afero.WriteFile(fs, "/catalog/service2.yml", []byte(`
upstream_services:
  - name: "service2"
    grpc_service:
      address: "grpc://service2"
`), 0644))
			},
			verify: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				assert.Len(t, services, 2)

				// Find service1
				var s1 *configv1.UpstreamServiceConfig
				for _, s := range services {
					if s.GetName() == "service1" {
						s1 = s
						break
					}
				}
				require.NotNil(t, s1, "service1 not found")
				require.NotNil(t, s1.GetHttpService(), "service1 should have http_service")
				assert.Equal(t, "http://service1", s1.GetHttpService().GetAddress())

				// Find service2
				var s2 *configv1.UpstreamServiceConfig
				for _, s := range services {
					if s.GetName() == "service2" {
						s2 = s
						break
					}
				}
				require.NotNil(t, s2, "service2 not found")
				require.NotNil(t, s2.GetGrpcService(), "service2 should have grpc_service")
				assert.Equal(t, "grpc://service2", s2.GetGrpcService().GetAddress())
			},
			expectedErr: false,
		},
		{
			name: "Empty Directory",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/catalog", 0755))
			},
			verify: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				assert.Empty(t, services)
			},
			expectedErr: false,
		},
		{
			name: "Directory Not Found",
			setupFs: func(fs afero.Fs) {
				// Do not create directory
			},
			verify:      nil,
			expectedErr: true,
		},
		{
			name: "Invalid YAML - Log and Continue",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/catalog", 0755))
				require.NoError(t, afero.WriteFile(fs, "/catalog/valid.yaml", []byte(`
upstream_services:
  - name: "valid"
    http_service:
      address: "http://valid"
`), 0644))
				require.NoError(t, afero.WriteFile(fs, "/catalog/invalid.yaml", []byte(`invalid: [ unclosed`), 0644))
			},
			verify: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				assert.Len(t, services, 1)
				assert.Equal(t, "valid", services[0].GetName())
			},
			expectedErr: false,
		},
		{
			name: "Non-YAML Files - Ignored",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/catalog", 0755))
				require.NoError(t, afero.WriteFile(fs, "/catalog/service.txt", []byte(`some text`), 0644))
			},
			verify: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				assert.Empty(t, services)
			},
			expectedErr: false,
		},
		{
			name: "Subdirectories - Scanned",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/catalog/subdir", 0755))
				require.NoError(t, afero.WriteFile(fs, "/catalog/subdir/service.yaml", []byte(`
upstream_services:
  - name: "nested"
    http_service:
      address: "http://nested"
`), 0644))
			},
			verify: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				assert.Len(t, services, 1)
				assert.Equal(t, "nested", services[0].GetName())
			},
			expectedErr: false,
		},
		{
			name: "Empty File - Should not Panic",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/catalog", 0755))
				require.NoError(t, afero.WriteFile(fs, "/catalog/empty.yaml", []byte(""), 0644))
			},
			verify: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				assert.Empty(t, services)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			m := NewManager(fs, "/catalog")
			err := m.Load(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !tt.expectedErr {
				services, err := m.ListServices(context.Background())
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t, services)
				}
			}
		})
	}
}

func TestManager_Concurrency(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("/catalog", 0755))
	require.NoError(t, afero.WriteFile(fs, "/catalog/s1.yaml", []byte(`upstream_services: [{name: s1, http_service: {address: "http://s1"}}]`), 0644))

	m := NewManager(fs, "/catalog")
	require.NoError(t, m.Load(context.Background()))

	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_ = m.Load(context.Background())
			}
		}
	}()

	for i := 0; i < 100; i++ {
		services, err := m.ListServices(context.Background())
		assert.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "s1", services[0].GetName())
	}
}
