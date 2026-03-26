// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStore_YAML_Tab_Error ...
// Summary: TestStore_YAML_Tab_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	// Create a YAML file with a tab character
	content := []byte("global_settings:\n\tlog_level: debug")
	err := afero.WriteFile(fs, "/config.yaml", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.yaml"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YAML files cannot contain tabs")
}

// TestStore_ClaudeDesktop_Error ...
// Summary: TestStore_ClaudeDesktop_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	content := []byte(`
mcpServers:
  weather:
    command: npx
`)
	err := afero.WriteFile(fs, "/config.yaml", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.yaml"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Claude Desktop configuration format")
}

// TestStore_Services_Alias_Error ...
// Summary: TestStore_Services_Alias_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	content := []byte(`
services:
  - name: weather
`)
	err := afero.WriteFile(fs, "/config.yaml", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.yaml"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "\"services\" is not a valid top-level key")
}

// TestStore_ServiceConfig_Wrapper_Error ...
// Summary: TestStore_ServiceConfig_Wrapper_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	content := []byte(`
upstream_services:
  - name: weather
    service_config:
      command_line_service:
        command: date
`)
	err := afero.WriteFile(fs, "/config.yaml", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.yaml"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "using 'service_config' as a wrapper key")
}

// TestStore_UnknownField_Suggestion ...
// Summary: TestStore_UnknownField_Suggestion
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	content := []byte(`
global_settings:
  mcp_litsen_address: :8080
`)
	err := afero.WriteFile(fs, "/config.yaml", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.yaml"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Did you mean \"mcp_listen_address\"?")
}

// TestStore_JSON_Error_Suggestions ...
// Summary: TestStore_JSON_Error_Suggestions
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	content := []byte(`{
		"mcpServers": {}
	}`)
	err := afero.WriteFile(fs, "/config.json", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.json"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Claude Desktop configuration format")
}

// TestStore_JSON_UnknownField ...
// Summary: TestStore_JSON_UnknownField
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	fs := afero.NewMemMapFs()
	content := []byte(`{
		"global_settings": {
			"mcp_litsen_address": ":8080"
		}
	}`)
	err := afero.WriteFile(fs, "/config.json", content, 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.json"})
	_, err = store.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Did you mean \"mcp_listen_address\"?")
}
