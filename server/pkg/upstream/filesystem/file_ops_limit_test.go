// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFileTool_SizeLimit(t *testing.T) {
	// Create a new in-memory filesystem provider
	prov := provider.NewTmpfsProvider()
	fs := prov.GetFs()

	// Create a file larger than 10MB
	const maxFileSize = 10 * 1024 * 1024
	// We don't need to actually allocate 10MB+1 bytes in memory if we can mock Stat,
	// but afero.MemMapFs stores everything in memory anyway.
	// 10MB is small enough for modern machines.
	largeContent := make([]byte, maxFileSize+1)
	err := afero.WriteFile(fs, "/large_file.txt", largeContent, 0644)
	require.NoError(t, err)

	// Get the tool definition
	toolDef := readFileTool(prov, fs)
	handler := toolDef.Handler

	// Execute the tool
	_, err = handler(context.Background(), map[string]interface{}{
		"path": "/large_file.txt",
	})

	// Assert error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file size exceeds limit")
}

func TestReadFileTool_Directory(t *testing.T) {
	// Create a new in-memory filesystem provider
	prov := provider.NewTmpfsProvider()
	fs := prov.GetFs()

	// Create a directory
	err := fs.MkdirAll("/mydir", 0755)
	require.NoError(t, err)

	// Get the tool definition
	toolDef := readFileTool(prov, fs)
	handler := toolDef.Handler

	// Execute the tool
	_, err = handler(context.Background(), map[string]interface{}{
		"path": "/mydir",
	})

	// Assert error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is a directory")
}

func TestReadFileTool_Success(t *testing.T) {
	// Create a new in-memory filesystem provider
	prov := provider.NewTmpfsProvider()
	fs := prov.GetFs()

	// Create a small file
	content := "hello world"
	err := afero.WriteFile(fs, "/hello.txt", []byte(content), 0644)
	require.NoError(t, err)

	// Get the tool definition
	toolDef := readFileTool(prov, fs)
	handler := toolDef.Handler

	// Execute the tool
	result, err := handler(context.Background(), map[string]interface{}{
		"path": "/hello.txt",
	})

	// Assert success
	require.NoError(t, err)
	assert.Equal(t, content, result["content"])
}
