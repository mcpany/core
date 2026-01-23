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

func TestLoadRegexTrimConsistency(t *testing.T) {
	// Tests that validation trims whitespace same as runtime resolution
	fs := afero.NewOsFs()
	// Create a temp config file
	configContent := `
upstream_services:
  - name: test-trim
    mcp_service:
      stdio_connection:
        command: ls
        env:
          TEST_TRIM:
            plain_text: "  secret  "
            validation_regex: "^secret$"
`
	tmpFile, err := afero.TempFile(fs, "", "config-*.yaml")
	require.NoError(t, err)
	defer fs.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Mock execLookPath to avoid dependency on system ls
	cleanup := mockExecLookPath()
	defer cleanup()

	store := NewFileStore(fs, []string{tmpFile.Name()})
	_, err = LoadServices(context.Background(), store, "server")

	// This should PASS if validation trims whitespace (consistent with runtime)
	assert.NoError(t, err)
}
