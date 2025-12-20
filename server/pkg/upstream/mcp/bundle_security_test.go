// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnzipBundle_DecompressionBomb(t *testing.T) {
	// Adjust maxUnzipSize for this test
	originalMaxUnzipSize := maxUnzipSize
	maxUnzipSize = 10 // Set limit to 10 bytes
	defer func() { maxUnzipSize = originalMaxUnzipSize }()

	// Create a zip with a file larger than 10 bytes
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("bomb.txt")
	require.NoError(t, err)
	// Write 20 bytes
	_, err = f.Write([]byte("01234567890123456789"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	// Try to unzip
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "bomb.zip")
	err = os.WriteFile(zipPath, buf.Bytes(), 0600)
	require.NoError(t, err)

	destDir := filepath.Join(tmpDir, "dest")
	err = unzipBundle(zipPath, destDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}
