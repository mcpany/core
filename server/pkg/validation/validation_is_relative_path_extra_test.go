// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/validation"
	"github.com/stretchr/testify/require"
)

func TestIsRelativePath_AllowedDotDotPrefix(t *testing.T) {
	// Create a temporary directory for our "CWD"
	cwd, err := os.MkdirTemp("", "mcpany-cwd")
	require.NoError(t, err)
	defer os.RemoveAll(cwd)

	// Change CWD to the temp dir
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(cwd)
	require.NoError(t, err)

	// Create a file named "..foo" in the CWD
	// This is a valid filename on most FS, and technically starts with ".." string
	// but is NOT a parent directory reference.
	err = os.Mkdir(filepath.Join(cwd, "..foo"), 0755)
	require.NoError(t, err)

	// This should PASS, but due to bug it FAILS
	err = validation.IsRelativePath("..foo")
	require.NoError(t, err, "IsRelativePath should allow filenames starting with '..'")
}
