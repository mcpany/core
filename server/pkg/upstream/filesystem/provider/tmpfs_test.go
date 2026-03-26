// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTmpfsProvider_Usage(t *testing.T) {
	p := NewTmpfsProvider()
	defer p.Close()

	// 1. Verify filesystem type
	fs := p.GetFs()
	assert.NotNil(t, fs)
	_, isMem := fs.(*afero.MemMapFs)
	assert.True(t, isMem, "Expected MemMapFs")

	// 2. Perform file operations
	fname := "test.txt"
	content := []byte("hello world")

	// Write
	err := afero.WriteFile(fs, fname, content, 0644)
	require.NoError(t, err)

	// Read
	readContent, err := afero.ReadFile(fs, fname)
	require.NoError(t, err)
	assert.Equal(t, content, readContent)

	// 3. ResolvePath
	resolved, err := p.ResolvePath(fname)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Clean(fname), resolved)

	resolved, err = p.ResolvePath("a/../b")
	assert.NoError(t, err)
	assert.Equal(t, "b", resolved)
}

func TestTmpfsProvider_Isolation(t *testing.T) {
	p1 := NewTmpfsProvider()
	defer p1.Close()

	p2 := NewTmpfsProvider()
	defer p2.Close()

	// Write to p1
	err := afero.WriteFile(p1.GetFs(), "file1", []byte("data"), 0644)
	require.NoError(t, err)

	// Check p2
	exists, err := afero.Exists(p2.GetFs(), "file1")
	assert.NoError(t, err)
	assert.False(t, exists, "p2 should not see files from p1")
}
