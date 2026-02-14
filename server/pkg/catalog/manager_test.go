// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Load_Parallel(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create many files to simulate load
	count := 50
	for i := 0; i < count; i++ {
		content := fmt.Sprintf(`
upstream_services:
  - name: service-%d
    http_service:
      address: http://service-%d.local
`, i, i)
		fname := fmt.Sprintf("/catalog/service-%d/config.yaml", i)
		require.NoError(t, fs.MkdirAll(fmt.Sprintf("/catalog/service-%d", i), 0755))
		require.NoError(t, afero.WriteFile(fs, fname, []byte(content), 0644))
	}

	m := NewManager(fs, "/catalog")
	err := m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, count)

	// Verify all services are present
	visited := make(map[string]bool)
	for _, svc := range services {
		visited[svc.GetName()] = true
	}
	assert.Len(t, visited, count)
	for i := 0; i < count; i++ {
		assert.True(t, visited[fmt.Sprintf("service-%d", i)], "service-%d missing", i)
	}
}

func TestManager_Load_MultipleServicesPerFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	content := `
upstream_services:
  - name: service-a
    http_service:
      address: http://a.local
  - name: service-b
    http_service:
      address: http://b.local
`
	require.NoError(t, fs.MkdirAll("/catalog/multi", 0755))
	require.NoError(t, afero.WriteFile(fs, "/catalog/multi/config.yaml", []byte(content), 0644))

	m := NewManager(fs, "/catalog")
	err := m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, 2)

	names := []string{services[0].GetName(), services[1].GetName()}
	sort.Strings(names)
	assert.Equal(t, []string{"service-a", "service-b"}, names)
}
