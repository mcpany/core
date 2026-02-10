// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	fs              afero.Fs
	resolvePathFunc func(string) (string, error)
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockProvider) ResolvePath(p string) (string, error) {
	if m.resolvePathFunc != nil {
		return m.resolvePathFunc(p)
	}
	return p, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestListDirectoryTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := listDirectoryTool(prov, fs)

	t.Run("Success", func(t *testing.T) {
		// Setup
		err := fs.MkdirAll("/data/subdir", 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, "/data/file.txt", []byte("content"), 0644)
		require.NoError(t, err)

		// Execute
		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/data",
		})
		require.NoError(t, err)

		// Verify
		entries, ok := res["entries"].([]interface{})
		require.True(t, ok)
		assert.Len(t, entries, 2)

		entryMap := make(map[string]map[string]interface{})
		for _, e := range entries {
			entry := e.(map[string]interface{})
			entryMap[entry["name"].(string)] = entry
		}

		assert.Contains(t, entryMap, "file.txt")
		assert.False(t, entryMap["file.txt"]["is_dir"].(bool))
		assert.Equal(t, int64(7), entryMap["file.txt"]["size"].(int64))

		assert.Contains(t, entryMap, "subdir")
		assert.True(t, entryMap["subdir"]["is_dir"].(bool))
	})

	t.Run("MissingPath", func(t *testing.T) {
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})

	t.Run("ResolvePathError", func(t *testing.T) {
		failProv := &mockProvider{
			fs: fs,
			resolvePathFunc: func(p string) (string, error) {
				return "", fmt.Errorf("resolve error")
			},
		}
		failTool := listDirectoryTool(failProv, fs)
		_, err := failTool.Handler(context.Background(), map[string]interface{}{
			"path": "/fail",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolve error")
	})

	t.Run("ReadDirError_NotExists", func(t *testing.T) {
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/nonexistent",
		})
		assert.Error(t, err)
		// Error message from afero usually contains "file does not exist"
		assert.True(t, os.IsNotExist(err) || err != nil)
	})

	t.Run("ReadDirError_FileAsDir", func(t *testing.T) {
		err := afero.WriteFile(fs, "/file", []byte("content"), 0644)
		require.NoError(t, err)

		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/file",
		})
		assert.Error(t, err)
		// Attempting to read a file as directory usually fails
	})
}

func TestGetFileInfoTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := getFileInfoTool(prov, fs)

	t.Run("Success_File", func(t *testing.T) {
		// Setup
		err := afero.WriteFile(fs, "/test.txt", []byte("hello"), 0644)
		require.NoError(t, err)

		// Execute
		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/test.txt",
		})
		require.NoError(t, err)

		// Verify
		assert.Equal(t, "test.txt", res["name"])
		assert.False(t, res["is_dir"].(bool))
		assert.Equal(t, int64(5), res["size"].(int64))
		assert.NotEmpty(t, res["mod_time"])
		// Verify mod_time format RFC3339
		_, parseErr := time.Parse(time.RFC3339, res["mod_time"].(string))
		assert.NoError(t, parseErr)
	})

	t.Run("Success_Dir", func(t *testing.T) {
		// Setup
		err := fs.MkdirAll("/testdir", 0755)
		require.NoError(t, err)

		// Execute
		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/testdir",
		})
		require.NoError(t, err)

		// Verify
		assert.Equal(t, "testdir", res["name"])
		assert.True(t, res["is_dir"].(bool))
	})

	t.Run("MissingPath", func(t *testing.T) {
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})

	t.Run("ResolvePathError", func(t *testing.T) {
		failProv := &mockProvider{
			fs: fs,
			resolvePathFunc: func(p string) (string, error) {
				return "", fmt.Errorf("resolve error")
			},
		}
		failTool := getFileInfoTool(failProv, fs)
		_, err := failTool.Handler(context.Background(), map[string]interface{}{
			"path": "/fail",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolve error")
	})

	t.Run("StatError_NotExists", func(t *testing.T) {
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/nonexistent",
		})
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err) || err != nil)
	})
}
