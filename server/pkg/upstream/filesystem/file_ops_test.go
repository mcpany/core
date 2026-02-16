// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMockProvider is a mock implementation of provider.Provider
type TestMockProvider struct {
	mock.Mock
}

var _ provider.Provider = (*TestMockProvider)(nil)

func (m *TestMockProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *TestMockProvider) GetFs() afero.Fs {
	args := m.Called()
	return args.Get(0).(afero.Fs)
}

func (m *TestMockProvider) ResolvePath(virtualPath string) (string, error) {
	args := m.Called(virtualPath)
	return args.String(0), args.Error(1)
}

func TestReadFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()

	t.Run("success", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/test.txt").Return("/test.txt", nil)

		toolDef := readFileTool(mockProv, fs)
		handler := toolDef.Handler

		err := afero.WriteFile(fs, "/test.txt", []byte("hello world"), 0644)
		assert.NoError(t, err)

		result, err := handler(context.Background(), map[string]interface{}{
			"path": "/test.txt",
		})
		assert.NoError(t, err)
		assert.Equal(t, "hello world", result["content"])
		mockProv.AssertExpectations(t)
	})

	t.Run("file_not_found", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/nonexistent.txt").Return("/nonexistent.txt", nil)

		toolDef := readFileTool(mockProv, fs)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/nonexistent.txt",
		})
		assert.Error(t, err)
		// Error message might vary depending on OS, checking for common part
		assert.Contains(t, err.Error(), "file does not exist")
		mockProv.AssertExpectations(t)
	})

	t.Run("is_directory", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/testdir").Return("/testdir", nil)

		toolDef := readFileTool(mockProv, fs)
		handler := toolDef.Handler

		err := fs.Mkdir("/testdir", 0755)
		assert.NoError(t, err)

		_, err = handler(context.Background(), map[string]interface{}{
			"path": "/testdir",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is a directory")
		mockProv.AssertExpectations(t)
	})

	t.Run("file_too_large", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/large.txt").Return("/large.txt", nil)

		toolDef := readFileTool(mockProv, fs)
		handler := toolDef.Handler

		// Create a file larger than 10MB
		size := 10*1024*1024 + 1
		largeContent := make([]byte, size)
		err := afero.WriteFile(fs, "/large.txt", largeContent, 0644)
		assert.NoError(t, err)

		_, err = handler(context.Background(), map[string]interface{}{
			"path": "/large.txt",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file size exceeds limit")
		mockProv.AssertExpectations(t)
	})

	t.Run("path_required", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		toolDef := readFileTool(mockProv, fs)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})
}

func TestWriteFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()

	t.Run("success", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/new.txt").Return("/new.txt", nil)

		toolDef := writeFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		result, err := handler(context.Background(), map[string]interface{}{
			"path":    "/new.txt",
			"content": "new content",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])

		// Verify file content
		content, err := afero.ReadFile(fs, "/new.txt")
		assert.NoError(t, err)
		assert.Equal(t, "new content", string(content))
		mockProv.AssertExpectations(t)
	})

	t.Run("nested_path", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/a/b/c/nested.txt").Return("/a/b/c/nested.txt", nil)

		toolDef := writeFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		result, err := handler(context.Background(), map[string]interface{}{
			"path":    "/a/b/c/nested.txt",
			"content": "nested content",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])

		content, err := afero.ReadFile(fs, "/a/b/c/nested.txt")
		assert.NoError(t, err)
		assert.Equal(t, "nested content", string(content))
		mockProv.AssertExpectations(t)
	})

	t.Run("read_only", func(t *testing.T) {
		mockProv := new(TestMockProvider)

		toolDef := writeFileTool(mockProv, fs, true)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{
			"path":    "/readonly.txt",
			"content": "fail",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	t.Run("content_required", func(t *testing.T) {
		mockProv := new(TestMockProvider)

		toolDef := writeFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/no_content.txt",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content is required")
	})
}

func TestMoveFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()

	t.Run("success_rename", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/source.txt").Return("/source.txt", nil)
		mockProv.On("ResolvePath", "/dest.txt").Return("/dest.txt", nil)

		toolDef := moveFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		err := afero.WriteFile(fs, "/source.txt", []byte("content"), 0644)
		assert.NoError(t, err)

		result, err := handler(context.Background(), map[string]interface{}{
			"source":      "/source.txt",
			"destination": "/dest.txt",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])

		// Verify source gone
		_, err = fs.Stat("/source.txt")
		assert.True(t, parserIsNotExist(err) || err != nil) // afero uses different errors sometimes

		// Verify dest exists
		content, err := afero.ReadFile(fs, "/dest.txt")
		assert.NoError(t, err)
		assert.Equal(t, "content", string(content))

		mockProv.AssertExpectations(t)
	})

	t.Run("success_move_nested", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/source_nested.txt").Return("/source_nested.txt", nil)
		mockProv.On("ResolvePath", "/x/y/z/dest.txt").Return("/x/y/z/dest.txt", nil)

		toolDef := moveFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		err := afero.WriteFile(fs, "/source_nested.txt", []byte("nested"), 0644)
		assert.NoError(t, err)

		result, err := handler(context.Background(), map[string]interface{}{
			"source":      "/source_nested.txt",
			"destination": "/x/y/z/dest.txt",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])

		_, err = fs.Stat("/x/y/z/dest.txt")
		assert.NoError(t, err)

		mockProv.AssertExpectations(t)
	})

	t.Run("read_only", func(t *testing.T) {
		mockProv := new(TestMockProvider)

		toolDef := moveFileTool(mockProv, fs, true)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{
			"source":      "/src",
			"destination": "/dst",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	t.Run("source_not_found", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/missing.txt").Return("/missing.txt", nil)
		mockProv.On("ResolvePath", "/dest.txt").Return("/dest.txt", nil)

		toolDef := moveFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{
			"source":      "/missing.txt",
			"destination": "/dest.txt",
		})
		assert.Error(t, err)
		mockProv.AssertExpectations(t)
	})
}

func TestDeleteFileTool(t *testing.T) {
	// Use OS filesystem for delete tests to ensure proper behavior with non-empty directories
	// afero.MemMapFs allows Remove() on non-empty directories which is not standard os.Remove behavior.
	tempDir := t.TempDir()
	fs := afero.NewBasePathFs(afero.NewOsFs(), tempDir)

	t.Run("success_file", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/to_delete.txt").Return("/to_delete.txt", nil)

		toolDef := deleteFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		err := afero.WriteFile(fs, "/to_delete.txt", []byte("bye"), 0644)
		assert.NoError(t, err)

		result, err := handler(context.Background(), map[string]interface{}{
			"path": "/to_delete.txt",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])

		_, err = fs.Stat("/to_delete.txt")
		assert.True(t, parserIsNotExist(err) || err != nil)
		mockProv.AssertExpectations(t)
	})

	t.Run("success_recursive_dir", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/dir_delete").Return("/dir_delete", nil)

		toolDef := deleteFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		err := fs.MkdirAll("/dir_delete/subdir", 0755)
		assert.NoError(t, err)
		err = afero.WriteFile(fs, "/dir_delete/subdir/file.txt", []byte("content"), 0644)
		assert.NoError(t, err)

		result, err := handler(context.Background(), map[string]interface{}{
			"path":      "/dir_delete",
			"recursive": true,
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])

		_, err = fs.Stat("/dir_delete")
		assert.True(t, parserIsNotExist(err) || err != nil)
		mockProv.AssertExpectations(t)
	})

	t.Run("fail_non_recursive_dir", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		mockProv.On("ResolvePath", "/dir_fail").Return("/dir_fail", nil)

		toolDef := deleteFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		err := fs.MkdirAll("/dir_fail/subdir", 0755)
		assert.NoError(t, err)
		// Need a file inside too to be sure it's not empty?
		// Usually os.Remove fails on non-empty dir regardless of files or dirs inside.
		err = afero.WriteFile(fs, "/dir_fail/subdir/file.txt", []byte("content"), 0644)
		assert.NoError(t, err)

		_, err = handler(context.Background(), map[string]interface{}{
			"path":      "/dir_fail",
			"recursive": false,
		})
		assert.Error(t, err)
		mockProv.AssertExpectations(t)
	})

	t.Run("read_only", func(t *testing.T) {
		mockProv := new(TestMockProvider)

		toolDef := deleteFileTool(mockProv, fs, true)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{
			"path": "/any",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem is read-only")
	})

	t.Run("path_required", func(t *testing.T) {
		mockProv := new(TestMockProvider)
		toolDef := deleteFileTool(mockProv, fs, false)
		handler := toolDef.Handler

		_, err := handler(context.Background(), map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})
}

// Helper to handle afero/os IsNotExist variance
func parserIsNotExist(err error) bool {
	if err == nil {
		return false
	}
	// afero memory fs might return different errors
	return true
}
