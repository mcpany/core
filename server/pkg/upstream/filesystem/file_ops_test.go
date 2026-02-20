// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FileOpsMockProvider implements the provider.Provider interface for testing.
// Using a unique name to avoid conflict with other test files in the package.
type FileOpsMockProvider struct {
	fs              afero.Fs
	resolvePathFunc func(virtualPath string) (string, error)
}

func (m *FileOpsMockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *FileOpsMockProvider) ResolvePath(virtualPath string) (string, error) {
	if m.resolvePathFunc != nil {
		return m.resolvePathFunc(virtualPath)
	}
	return virtualPath, nil
}

func (m *FileOpsMockProvider) Close() error {
	return nil
}

// Helper to create a temp FS for testing
func createTempFs(t *testing.T) (afero.Fs, func()) {
	dir, err := os.MkdirTemp("", "fs_test")
	require.NoError(t, err)

	fs := afero.NewBasePathFs(afero.NewOsFs(), dir)

	cleanup := func() {
		os.RemoveAll(dir)
	}
	return fs, cleanup
}

func TestReadFileTool(t *testing.T) {
	fs, cleanup := createTempFs(t)
	defer cleanup()

	prov := &FileOpsMockProvider{fs: fs}
	toolDef := readFileTool(prov, fs)

	t.Run("Success", func(t *testing.T) {
		err := afero.WriteFile(fs, "/test.txt", []byte("hello world"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		res, err := handler(context.Background(), map[string]interface{}{
			"path": "/test.txt",
		})
		require.NoError(t, err)
		assert.Equal(t, "hello world", res["content"])
	})

	t.Run("Path Resolution Error", func(t *testing.T) {
		prov.resolvePathFunc = func(path string) (string, error) {
			return "", fmt.Errorf("resolution failed")
		}
		defer func() { prov.resolvePathFunc = nil }()

		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/bad.txt",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolution failed")
	})

	t.Run("Path Is Directory", func(t *testing.T) {
		err := fs.MkdirAll("/mydir", 0755)
		require.NoError(t, err)

		handler := toolDef.Handler
		_, err = handler(context.Background(), map[string]interface{}{
			"path": "/mydir",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is a directory")
	})

	t.Run("File Not Found", func(t *testing.T) {
		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/nonexistent.txt",
		})
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err) || strings.Contains(err.Error(), "no such file"))
	})

	t.Run("File Size Limit", func(t *testing.T) {
		largeFile := "/large.txt"
		f, err := fs.Create(largeFile)
		require.NoError(t, err)

		// Use package-level constant + 1 byte
		size := int64(maxFileSize + 1)

		// Seek to the end-1 and write a byte to make it sparse if supported, or just full size
		_, err = f.Seek(size-1, 0)
		require.NoError(t, err)
		_, err = f.Write([]byte{0})
		require.NoError(t, err)
		f.Close()

		handler := toolDef.Handler
		_, err = handler(context.Background(), map[string]interface{}{
			"path": largeFile,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file size exceeds limit")
	})
}

func TestWriteFileTool(t *testing.T) {
	fs, cleanup := createTempFs(t)
	defer cleanup()

	prov := &FileOpsMockProvider{fs: fs}

	// Test read-only flag
	t.Run("ReadOnly", func(t *testing.T) {
		toolDef := writeFileTool(prov, fs, true)
		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"path":    "/test.txt",
			"content": "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	toolDef := writeFileTool(prov, fs, false)

	t.Run("Success", func(t *testing.T) {
		handler := toolDef.Handler
		res, err := handler(context.Background(), map[string]interface{}{
			"path":    "/newfile.txt",
			"content": "new content",
		})
		require.NoError(t, err)
		assert.Equal(t, true, res["success"])

		content, err := afero.ReadFile(fs, "/newfile.txt")
		require.NoError(t, err)
		assert.Equal(t, "new content", string(content))
	})

	t.Run("Creates Parent Directories", func(t *testing.T) {
		handler := toolDef.Handler
		res, err := handler(context.Background(), map[string]interface{}{
			"path":    "/a/b/c/nested.txt",
			"content": "nested",
		})
		require.NoError(t, err)
		assert.Equal(t, true, res["success"])

		content, err := afero.ReadFile(fs, "/a/b/c/nested.txt")
		require.NoError(t, err)
		assert.Equal(t, "nested", string(content))
	})

	t.Run("Path Resolution Error", func(t *testing.T) {
		prov.resolvePathFunc = func(path string) (string, error) {
			return "", fmt.Errorf("resolution failed")
		}
		defer func() { prov.resolvePathFunc = nil }()

		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"path":    "/bad.txt",
			"content": "fail",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolution failed")
	})

	t.Run("Parent Creation Failure (File as Dir)", func(t *testing.T) {
		// specific subtest cleanup if needed, but main cleanup handles it

		// Create a file
		err := afero.WriteFile(fs, "/file_blocker", []byte("i am a file"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		// Try to create /file_blocker/child.txt
		_, err = handler(context.Background(), map[string]interface{}{
			"path":    "/file_blocker/child.txt",
			"content": "fail",
		})
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create parent directory")
		}
	})
}

func TestMoveFileTool(t *testing.T) {
	fs, cleanup := createTempFs(t)
	defer cleanup()

	prov := &FileOpsMockProvider{fs: fs}

	t.Run("ReadOnly", func(t *testing.T) {
		toolDef := moveFileTool(prov, fs, true)
		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"source":      "/src",
			"destination": "/dest",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	toolDef := moveFileTool(prov, fs, false)

	t.Run("Success", func(t *testing.T) {
		err := afero.WriteFile(fs, "/source.txt", []byte("content"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		res, err := handler(context.Background(), map[string]interface{}{
			"source":      "/source.txt",
			"destination": "/dest.txt",
		})
		require.NoError(t, err)
		assert.Equal(t, true, res["success"])

		// Check dest exists
		content, err := afero.ReadFile(fs, "/dest.txt")
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))

		// Check source gone
		exists, err := afero.Exists(fs, "/source.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Creates Parent Dirs", func(t *testing.T) {
		err := afero.WriteFile(fs, "/src_nested.txt", []byte("nested"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		_, err = handler(context.Background(), map[string]interface{}{
			"source":      "/src_nested.txt",
			"destination": "/x/y/z/dest_nested.txt",
		})
		require.NoError(t, err)

		content, err := afero.ReadFile(fs, "/x/y/z/dest_nested.txt")
		require.NoError(t, err)
		assert.Equal(t, "nested", string(content))
	})

	t.Run("Source Not Found", func(t *testing.T) {
		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"source":      "/missing.txt",
			"destination": "/dest.txt",
		})
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err) || strings.Contains(err.Error(), "does not exist"))
	})

	t.Run("Resolution Error", func(t *testing.T) {
		prov.resolvePathFunc = func(path string) (string, error) {
			return "", fmt.Errorf("resolution failed")
		}
		defer func() { prov.resolvePathFunc = nil }()

		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"source":      "/src",
			"destination": "/dest",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolution failed")
	})
}

func TestDeleteFileTool(t *testing.T) {
	fs, cleanup := createTempFs(t)
	defer cleanup()

	prov := &FileOpsMockProvider{fs: fs}

	t.Run("ReadOnly", func(t *testing.T) {
		toolDef := deleteFileTool(prov, fs, true)
		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/file",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	toolDef := deleteFileTool(prov, fs, false)

	t.Run("Success Recursive", func(t *testing.T) {
		err := fs.MkdirAll("/dir/subdir", 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, "/dir/file.txt", []byte("file"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		res, err := handler(context.Background(), map[string]interface{}{
			"path":      "/dir",
			"recursive": true,
		})
		require.NoError(t, err)
		assert.Equal(t, true, res["success"])

		exists, err := afero.Exists(fs, "/dir")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Success Non-Recursive File", func(t *testing.T) {
		err := afero.WriteFile(fs, "/single.txt", []byte("single"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		res, err := handler(context.Background(), map[string]interface{}{
			"path":      "/single.txt",
			"recursive": false,
		})
		require.NoError(t, err)
		assert.Equal(t, true, res["success"])

		exists, err := afero.Exists(fs, "/single.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Fail Non-Recursive Directory Not Empty", func(t *testing.T) {
		// Prepare non-empty dir
		err := fs.MkdirAll("/nonempty", 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, "/nonempty/file.txt", []byte("content"), 0644)
		require.NoError(t, err)

		handler := toolDef.Handler
		_, err = handler(context.Background(), map[string]interface{}{
			"path":      "/nonempty",
			"recursive": false,
		})
		// os.Remove on non-empty dir fails
		assert.Error(t, err)

		// Verify it still exists
		exists, _ := afero.Exists(fs, "/nonempty")
		assert.True(t, exists)
	})

	t.Run("Resolution Error", func(t *testing.T) {
		prov.resolvePathFunc = func(path string) (string, error) {
			return "", fmt.Errorf("resolution failed")
		}
		defer func() { prov.resolvePathFunc = nil }()

		handler := toolDef.Handler
		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/file",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolution failed")
	})
}
