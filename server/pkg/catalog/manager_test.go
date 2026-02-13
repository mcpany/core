// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Load(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create multiple dummy service files
	count := 20
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

	// Create manager
	m := NewManager(fs, "/catalog")
	err := m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, count)

	// Verify names
	loadedNames := make(map[string]bool)
	for _, svc := range services {
		loadedNames[svc.GetName()] = true
	}
	for i := 0; i < count; i++ {
		assert.True(t, loadedNames[fmt.Sprintf("service-%d", i)], "Missing service-%d", i)
	}
}
