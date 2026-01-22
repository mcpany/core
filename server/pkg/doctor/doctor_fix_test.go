// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunChecks_Filesystem_Fix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor-fix-fs")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Path to a non-existent directory
	missingDir := filepath.Join(tmpDir, "missing-dir")

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("invalid-fs-fixable"),
				ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
					FilesystemService: &configv1.FilesystemUpstreamService{
						RootPaths: map[string]string{
							"/data": missingDir,
						},
					},
				},
			},
		},
	}

	// 1. Run Checks -> Expect Error and Fix available
	results := RunChecks(context.Background(), config)

	require.Len(t, results, 1)
	assert.Equal(t, StatusError, results[0].Status)
	assert.NotNil(t, results[0].Fix, "Fix function should not be nil")
	assert.Contains(t, results[0].FixName, "Create missing directory")

	// 2. Execute Fix
	err = results[0].Fix()
	assert.NoError(t, err, "Fix execution should succeed")

	// 3. Verify directory creation
	_, err = os.Stat(missingDir)
	assert.NoError(t, err, "Directory should exist after fix")

	// 4. Run Checks again -> Expect OK
	resultsAgain := RunChecks(context.Background(), config)
	require.Len(t, resultsAgain, 1)
	assert.Equal(t, StatusOk, resultsAgain[0].Status, "Status should be OK after fix")
}
