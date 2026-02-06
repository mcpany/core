// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	fs afero.Fs
}

func (m *mockProvider) Close() error { return nil }
func (m *mockProvider) GetFs() afero.Fs { return m.fs }
func (m *mockProvider) ResolvePath(virtualPath string) (string, error) {
	return virtualPath, nil
}

func TestReadFileTool_SizeLimit(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create a file larger than 10MB
	// 10MB = 10 * 1024 * 1024 = 10485760 bytes
	// We create one byte larger
	largeContent := make([]byte, 10485760+1)
	err := afero.WriteFile(fs, "/large.txt", largeContent, 0644)
	require.NoError(t, err)

	tool := readFileTool(prov, fs)
	_, err = tool.Handler(context.Background(), map[string]interface{}{
		"path": "/large.txt",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file size exceeds limit")
}

func TestFileTools_ReadOnly(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	t.Run("write_file", func(t *testing.T) {
		tool := writeFileTool(prov, fs, true)
		_, err := tool.Handler(context.Background(), map[string]interface{}{
			"path":    "/test.txt",
			"content": "foo",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	t.Run("move_file", func(t *testing.T) {
		tool := moveFileTool(prov, fs, true)
		_, err := tool.Handler(context.Background(), map[string]interface{}{
			"source":      "/src",
			"destination": "/dst",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	t.Run("delete_file", func(t *testing.T) {
		tool := deleteFileTool(prov, fs, true)
		_, err := tool.Handler(context.Background(), map[string]interface{}{
			"path": "/test.txt",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})
}

func TestReadFileTool_Directory(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	err := fs.Mkdir("/testdir", 0755)
	require.NoError(t, err)

	tool := readFileTool(prov, fs)
	_, err = tool.Handler(context.Background(), map[string]interface{}{
		"path": "/testdir",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is a directory")
}
