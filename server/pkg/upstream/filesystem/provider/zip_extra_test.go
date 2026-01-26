// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestZipProvider_E2E(t *testing.T) {
	// Create a temporary zip file
	tmpDir := t.TempDir()

	// Allow tempDir for validation
	validation.SetAllowedPaths([]string{tmpDir})
	defer validation.SetAllowedPaths(nil)

	zipPath := filepath.Join(tmpDir, "test.zip")

	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)

	w := zip.NewWriter(zipFile)

	// Add a file
	f, err := w.Create("hello.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("world"))
	require.NoError(t, err)

	// Add a directory/file
	f, err = w.Create("dir/file.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("content"))
	require.NoError(t, err)

	require.NoError(t, w.Close())
	require.NoError(t, zipFile.Close())

	// Create ZipProvider
	config := configv1.ZipFs_builder{
		FilePath: proto.String(zipPath),
	}.Build()
	p, err := NewZipProvider(config)
	require.NoError(t, err)
	defer p.Close()

	// Test GetFs
	fs := p.GetFs()
	assert.NotNil(t, fs)

	// Test ResolvePath
	resolved, err := p.ResolvePath("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", resolved)

	// Test Open and Read
	file, err := fs.Open("hello.txt")
	require.NoError(t, err)
	content, err := io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, "world", string(content))
	file.Close()

	// Test nested file
	file, err = fs.Open("dir/file.txt")
	require.NoError(t, err)
	content, err = io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, "content", string(content))
	file.Close()

	// Test Stat
	info, err := fs.Stat("hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", info.Name())
	assert.Equal(t, int64(5), info.Size())

	// Test Failures
	_, err = NewZipProvider(configv1.ZipFs_builder{FilePath: proto.String("non_existent.zip")}.Build())
	assert.Error(t, err)
}
