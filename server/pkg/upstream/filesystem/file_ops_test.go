// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFileOpsProvider implements provider.Provider for testing.
type mockFileOpsProvider struct {
	fs          afero.Fs
	resolveFunc func(string) (string, error)
}

func (m *mockFileOpsProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockFileOpsProvider) ResolvePath(virtualPath string) (string, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(virtualPath)
	}
	// Default: assume virtual path is the resolved path for MemMapFs
	return virtualPath, nil
}

func (m *mockFileOpsProvider) Close() error {
	return nil
}

func TestReadFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockFileOpsProvider{fs: fs}

	// Setup: Create a file
	err := afero.WriteFile(fs, "/test/file.txt", []byte("hello world"), 0644)
	require.NoError(t, err)

	toolDef := readFileTool(prov, fs)

	tests := []struct {
		name      string
		args      map[string]interface{}
		want      map[string]interface{}
		wantErr   bool
		errString string
	}{
		{
			name: "happy path",
			args: map[string]interface{}{"path": "/test/file.txt"},
			want: map[string]interface{}{"content": "hello world"},
		},
		{
			name:      "missing path arg",
			args:      map[string]interface{}{},
			wantErr:   true,
			errString: "path is required",
		},
		{
			name:      "file not found",
			args:      map[string]interface{}{"path": "/test/missing.txt"},
			wantErr:   true,
			errString: "does not exist", // afero error message part
		},
		{
			name: "path is directory",
			args: map[string]interface{}{"path": "/test"}, // /test is implicitly created as parent dir of /test/file.txt
			// MemMapFs behavior on opening dir might differ slightly but Stat works.
			wantErr:   true,
			errString: "path is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toolDef.Handler(context.Background(), tt.args)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}

	t.Run("resolve path error", func(t *testing.T) {
		failProv := &mockFileOpsProvider{
			fs: fs,
			resolveFunc: func(p string) (string, error) {
				return "", fmt.Errorf("resolve failed")
			},
		}
		failTool := readFileTool(failProv, fs)
		_, err := failTool.Handler(context.Background(), map[string]interface{}{"path": "/any"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resolve failed")
	})

	t.Run("file size limit", func(t *testing.T) {
		// Create a large file > 10MB
		largeContent := make([]byte, 10*1024*1024+1)
		err := afero.WriteFile(fs, "/test/large.txt", largeContent, 0644)
		require.NoError(t, err)

		_, err = toolDef.Handler(context.Background(), map[string]interface{}{"path": "/test/large.txt"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file size exceeds limit")
	})
}

func TestWriteFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockFileOpsProvider{fs: fs}

	t.Run("happy path create new", func(t *testing.T) {
		toolDef := writeFileTool(prov, fs, false)
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/new/file.txt",
			"content": "new content",
		})
		require.NoError(t, err)

		content, err := afero.ReadFile(fs, "/new/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "new content", string(content))
	})

	t.Run("happy path overwrite", func(t *testing.T) {
		err := afero.WriteFile(fs, "/existing.txt", []byte("old"), 0644)
		require.NoError(t, err)

		toolDef := writeFileTool(prov, fs, false)
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/existing.txt",
			"content": "new",
		})
		require.NoError(t, err)

		content, err := afero.ReadFile(fs, "/existing.txt")
		require.NoError(t, err)
		assert.Equal(t, "new", string(content))
	})

	t.Run("read only", func(t *testing.T) {
		toolDef := writeFileTool(prov, fs, true)
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/ro.txt",
			"content": "data",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")
	})

	t.Run("missing args", func(t *testing.T) {
		toolDef := writeFileTool(prov, fs, false)
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{"path": "/foo"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "content is required")
	})

	t.Run("path is directory conflict", func(t *testing.T) {
		err := fs.MkdirAll("/dir", 0755)
		require.NoError(t, err)

		toolDef := writeFileTool(prov, fs, false)
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/dir",
			"content": "data",
		})
		// safeWriteFile -> OpenFile(O_EXCL) should fail if exists.
		// If O_EXCL fails, it checks if it's a directory or tries open for truncate.
		// OpenFile on directory usually fails on most OSs, MemMapFs behavior:
		// IsExist error returned by OpenFile with O_CREATE|O_EXCL.
		// Then falls through to open for truncate.
		// OpenFile("/dir", O_WRONLY) on MemMapFs returns IsDir error?
		require.Error(t, err)
	})
}

func TestMoveFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockFileOpsProvider{fs: fs}

	t.Run("happy path rename", func(t *testing.T) {
		err := afero.WriteFile(fs, "/src.txt", []byte("data"), 0644)
		require.NoError(t, err)

		toolDef := moveFileTool(prov, fs, false)
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"source":      "/src.txt",
			"destination": "/dest.txt",
		})
		require.NoError(t, err)

		exists, _ := afero.Exists(fs, "/src.txt")
		assert.False(t, exists)
		exists, _ = afero.Exists(fs, "/dest.txt")
		assert.True(t, exists)
	})

	t.Run("happy path move create parent", func(t *testing.T) {
		err := afero.WriteFile(fs, "/src2.txt", []byte("data"), 0644)
		require.NoError(t, err)

		toolDef := moveFileTool(prov, fs, false)
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"source":      "/src2.txt",
			"destination": "/a/b/dest.txt",
		})
		require.NoError(t, err)

		exists, _ := afero.Exists(fs, "/a/b/dest.txt")
		assert.True(t, exists)
	})

	t.Run("source not found", func(t *testing.T) {
		toolDef := moveFileTool(prov, fs, false)
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"source":      "/missing.txt",
			"destination": "/dest.txt",
		})
		require.Error(t, err)
	})

	t.Run("read only", func(t *testing.T) {
		toolDef := moveFileTool(prov, fs, true)
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"source":      "/src.txt",
			"destination": "/dest.txt",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")
	})
}

func TestDeleteFileTool(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockFileOpsProvider{fs: fs}

	t.Run("happy path file", func(t *testing.T) {
		err := afero.WriteFile(fs, "/del.txt", []byte("data"), 0644)
		require.NoError(t, err)

		toolDef := deleteFileTool(prov, fs, false)
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/del.txt",
		})
		require.NoError(t, err)

		exists, _ := afero.Exists(fs, "/del.txt")
		assert.False(t, exists)
	})

	t.Run("happy path recursive", func(t *testing.T) {
		err := fs.MkdirAll("/dir/sub", 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, "/dir/sub/file.txt", []byte("data"), 0644)
		require.NoError(t, err)

		toolDef := deleteFileTool(prov, fs, false)
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path":      "/dir",
			"recursive": true,
		})
		require.NoError(t, err)

		exists, _ := afero.Exists(fs, "/dir")
		assert.False(t, exists)
	})

	t.Run("non recursive dir fail", func(t *testing.T) {
		// Using OsFs because MemMapFs.Remove(dir) succeeds on non-empty dirs (bug/feature of MemMapFs).
		// We want to verify that safeRemove fails as expected on a real filesystem.
		realFs := afero.NewOsFs()
		tempDir := t.TempDir()
		realProv := &mockFileOpsProvider{fs: realFs, resolveFunc: func(p string) (string, error) {
			return filepath.Join(tempDir, p), nil
		}}

		dirPath := filepath.Join(tempDir, "dir2")
		subPath := filepath.Join(dirPath, "sub")
		err := realFs.MkdirAll(subPath, 0755)
		require.NoError(t, err)

		toolDef := deleteFileTool(realProv, realFs, false)
		// Passing relative path "dir2" which resolveFunc maps to tempDir/dir2
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path":      "dir2",
			"recursive": false,
		})
		require.Error(t, err)
	})

	t.Run("read only", func(t *testing.T) {
		toolDef := deleteFileTool(prov, fs, true)
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/foo",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")
	})
}

// verifyPathIntegrity is internal but critical. Since it relies on OS-specific behavior (EvalSymlinks),
// checking it with MemMapFs skips the logic (as per implementation).
// We rely on file_ops_security_test.go for that.
// But we can check that it doesn't crash on MemMapFs.

func TestVerifyPathIntegrity_MemMapFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := verifyPathIntegrity(fs, "/any/path")
	require.NoError(t, err)
}

func TestSafeWriteFile_MemMapFs_Simple(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := safeWriteFile(fs, "/test.txt", []byte("data"), 0644)
	require.NoError(t, err)

	content, _ := afero.ReadFile(fs, "/test.txt")
	assert.Equal(t, "data", string(content))
}
