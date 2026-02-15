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

// MockProvider implements provider.Provider for testing purposes.
type MockProvider struct {
	fs              afero.Fs
	resolvePathFunc func(string) (string, error)
}

func (m *MockProvider) GetFs() afero.Fs {
	if m.fs != nil {
		return m.fs
	}
	return afero.NewMemMapFs()
}

func (m *MockProvider) ResolvePath(virtualPath string) (string, error) {
	if m.resolvePathFunc != nil {
		return m.resolvePathFunc(virtualPath)
	}
	return virtualPath, nil
}

func (m *MockProvider) Close() error {
	return nil
}

func TestListDirectoryTool(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Setup filesystem
		fs := afero.NewMemMapFs()
		fs.MkdirAll("/data/subdir", 0755)
		afero.WriteFile(fs, "/data/file1.txt", []byte("content"), 0644)
		afero.WriteFile(fs, "/data/file2.log", []byte("log"), 0644)

		prov := &MockProvider{}
		toolDef := listDirectoryTool(prov, fs)

		// Execute
		res, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/data",
		})
		require.NoError(t, err)

		// Verify result structure
		resMap, ok := res["entries"].([]interface{})
		require.True(t, ok)
		assert.Len(t, resMap, 3) // subdir, file1.txt, file2.log

		// Verify contents
		names := make(map[string]bool)
		for _, entry := range resMap {
			e, ok := entry.(map[string]interface{})
			require.True(t, ok)
			names[e["name"].(string)] = true

			if e["name"] == "subdir" {
				assert.True(t, e["is_dir"].(bool))
			} else {
				assert.False(t, e["is_dir"].(bool))
			}
		}
		assert.True(t, names["subdir"])
		assert.True(t, names["file1.txt"])
		assert.True(t, names["file2.log"])
	})

	t.Run("MissingPath", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		prov := &MockProvider{}
		toolDef := listDirectoryTool(prov, fs)

		_, err := toolDef.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})

	t.Run("ResolvePathError", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		prov := &MockProvider{
			resolvePathFunc: func(p string) (string, error) {
				return "", fmt.Errorf("resolve error")
			},
		}
		toolDef := listDirectoryTool(prov, fs)

		_, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/restricted",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolve error")
	})

	t.Run("ReadDirError_NonExistent", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		prov := &MockProvider{}
		toolDef := listDirectoryTool(prov, fs)

		_, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/nonexistent",
		})
		assert.Error(t, err)
		// Error message depends on OS, but usually "file does not exist"
		assert.True(t, os.IsNotExist(err) || err != nil)
	})

	t.Run("ReadDirError_FileAsDir", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		afero.WriteFile(fs, "/file.txt", []byte("content"), 0644)
		prov := &MockProvider{}
		toolDef := listDirectoryTool(prov, fs)

		// Trying to list a file as a directory
		// afero.ReadDir might return specific error or empty list?
		// os.ReadDir returns error if path is not a directory.
		_, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/file.txt",
		})
		assert.Error(t, err)
	})
}

func TestGetFileInfoTool(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_File", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		content := []byte("hello world")
		afero.WriteFile(fs, "/test.txt", content, 0644)

		// Set mod time explicitly to verify it? MemMapFs sets it to now.
		// We can just verify it parses correctly.

		prov := &MockProvider{}
		toolDef := getFileInfoTool(prov, fs)

		res, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/test.txt",
		})
		require.NoError(t, err)

		assert.Equal(t, "test.txt", res["name"])
		assert.Equal(t, false, res["is_dir"])
		assert.Equal(t, int64(len(content)), res["size"])
		_, err = time.Parse(time.RFC3339, res["mod_time"].(string))
		assert.NoError(t, err)
	})

	t.Run("Success_Dir", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		fs.MkdirAll("/subdir", 0755)

		prov := &MockProvider{}
		toolDef := getFileInfoTool(prov, fs)

		res, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/subdir",
		})
		require.NoError(t, err)

		assert.Equal(t, "subdir", res["name"])
		assert.Equal(t, true, res["is_dir"])
	})

	t.Run("MissingPath", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		prov := &MockProvider{}
		toolDef := getFileInfoTool(prov, fs)

		_, err := toolDef.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})

	t.Run("ResolvePathError", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		prov := &MockProvider{
			resolvePathFunc: func(p string) (string, error) {
				return "", fmt.Errorf("resolve error")
			},
		}
		toolDef := getFileInfoTool(prov, fs)

		_, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/restricted",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resolve error")
	})

	t.Run("StatError_NonExistent", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		prov := &MockProvider{}
		toolDef := getFileInfoTool(prov, fs)

		_, err := toolDef.Handler(ctx, map[string]interface{}{
			"path": "/nonexistent",
		})
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err) || err != nil)
	})
}
