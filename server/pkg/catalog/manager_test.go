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

// TestManager_Load_Parallel ...
// Summary: TestManager_Load_Parallel
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	// Create many files to simulate load
	count := 100
	for i := 0; i < count; i++ {
		content := fmt.Sprintf(`
upstream_services:
  - name: service-%d
    http_service:
      address: http://service-%d.local
`, i, i)
		fname := fmt.Sprintf("/catalog/service-%d/config.yaml", i)
		require.NoError(t, afero.WriteFile(fs, fname, []byte(content), 0644))
	}

	m := NewManager(fs, "/catalog")
	err := m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, count)

	// Verify a few random services
	found := make(map[string]bool)
	for _, svc := range services {
		found[svc.GetName()] = true
	}
	assert.True(t, found["service-0"])
	assert.True(t, found["service-99"])
}

// TestManager_Load_Empty ...
// Summary: TestManager_Load_Empty
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("/catalog", 0755))
	m := NewManager(fs, "/catalog")
	err := m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Empty(t, services)
}

// TestManager_Load_PartialFailure ...
// Summary: TestManager_Load_PartialFailure
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// One file is valid, one is invalid YAML
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/catalog/valid.yaml", []byte(`
upstream_services:
  - name: valid-service
    http_service:
      address: http://valid.local
`), 0644))
	// Invalid YAML - should be skipped/logged but not fail Load
	require.NoError(t, afero.WriteFile(fs, "/catalog/invalid.yaml", []byte(`invalid: yaml: [`), 0644))

	m := NewManager(fs, "/catalog")
	err := m.Load(context.Background())
	require.NoError(t, err)

	services, err := m.ListServices(context.Background())
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "valid-service", services[0].GetName())
}
