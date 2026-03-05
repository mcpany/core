// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/spf13/afero"
)

func TestNewCatalogServer(t *testing.T) {
	fs := afero.NewMemMapFs()
	manager := catalog.NewManager(fs, "/tmp/catalog")

	server := NewCatalogServer(manager)

	if server == nil {
		t.Fatal("expected NewCatalogServer to return a non-nil server")
	}

	if server.manager != manager {
		t.Errorf("expected server manager to be %p, got %p", manager, server.manager)
	}
}

func TestCatalogServer_ListServices(t *testing.T) {
	// Setup the mock file system and the catalog manager.
	fs := afero.NewMemMapFs()

	// Create a mock config file in the catalog to be loaded by the manager.
	mockConfig := `
upstream_services:
  - name: test-service


`
	if err := afero.WriteFile(fs, "/tmp/catalog/test-config.yaml", []byte(mockConfig), 0644); err != nil {
		t.Fatalf("failed to write mock config: %v", err)
	}

	tests := []struct {
		name         string
		setupManager func() *catalog.Manager
		expectErr    bool
		expectCount  int
	}{
		{
			name: "Happy Path - Empty Catalog",
			setupManager: func() *catalog.Manager {
				emptyFs := afero.NewMemMapFs()
				m := catalog.NewManager(emptyFs, "/tmp/catalog")
				_ = m.Load(context.Background())
				return m
			},
			expectErr:   false,
			expectCount: 0,
		},
		{
			name: "Happy Path - With Services",
			setupManager: func() *catalog.Manager {
				m := catalog.NewManager(fs, "/tmp/catalog")
				err := m.Load(context.Background())
				if err != nil {
					t.Fatalf("failed to load manager: %v", err)
				}
				return m
			},
			expectErr:   false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setupManager()
			server := NewCatalogServer(m)

			resp, err := server.ListServices(context.Background(), apiv1.ListCatalogServicesRequest_builder{}.Build())

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if !tt.expectErr {
				if resp == nil {
					t.Fatal("expected non-nil response")
				}

				if len(resp.GetServices()) != tt.expectCount {
					t.Errorf("expected %d services, got %d", tt.expectCount, len(resp.GetServices()))
				}

				// If there are services, do a basic check on the first one.
				if tt.expectCount > 0 && len(resp.GetServices()) > 0 {
					if resp.GetServices()[0].GetName() != "test-service" {
						t.Errorf("expected service name 'test-service', got '%s'", resp.GetServices()[0].GetName())
					}
				}
			}
		})
	}
}
