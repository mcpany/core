// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	fs afero.Fs
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockProvider) ResolvePath(virtualPath string) (string, error) {
	return virtualPath, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFilesTool(t *testing.T) {
	// Helper to setup and execute the tool
	setupAndExec := func(ctx context.Context, args map[string]interface{}, setup func(fs afero.Fs)) (map[string]interface{}, error) {
		fs := afero.NewMemMapFs()
		prov := &mockProvider{fs: fs}
		toolDef := searchFilesTool(prov, fs)
		if setup != nil {
			setup(fs)
		}
		return toolDef.Handler(ctx, args)
	}

	t.Run("HappyPath", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "hello",
		}, func(fs afero.Fs) {
			require.NoError(t, afero.WriteFile(fs, "/root/file1.txt", []byte("hello world\nfoo bar\nhello universe"), 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 2)

		assert.Equal(t, "/root/file1.txt", matches[0]["file"])
		assert.Equal(t, 1, matches[0]["line_number"])
		assert.Equal(t, "hello world", matches[0]["line_content"])

		assert.Equal(t, "/root/file1.txt", matches[1]["file"])
		assert.Equal(t, 3, matches[1]["line_number"])
		assert.Equal(t, "hello universe", matches[1]["line_content"])
	})

	t.Run("RegexSupport", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "\\d+",
		}, func(fs afero.Fs) {
			require.NoError(t, afero.WriteFile(fs, "/root/numbers.txt", []byte("abc 123 xyz\nno numbers here"), 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "abc 123 xyz", matches[0]["line_content"])
	})

	t.Run("InvalidRegex", func(t *testing.T) {
		_, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "[",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("ExclusionPatterns", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":             "/root",
			"pattern":          "match",
			"exclude_patterns": []interface{}{"*.test.js"},
		}, func(fs afero.Fs) {
			require.NoError(t, afero.WriteFile(fs, "/root/include.txt", []byte("match me"), 0644))
			require.NoError(t, afero.WriteFile(fs, "/root/exclude.test.js", []byte("match me"), 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/root/include.txt", matches[0]["file"])
	})

	t.Run("HiddenDirectories", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "match",
		}, func(fs afero.Fs) {
			// Hidden directory
			require.NoError(t, fs.MkdirAll("/root/.git", 0755))
			require.NoError(t, afero.WriteFile(fs, "/root/.git/config", []byte("match me"), 0644))

			// Visible directory
			require.NoError(t, afero.WriteFile(fs, "/root/visible/file.txt", []byte("match me"), 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/root/visible/file.txt", matches[0]["file"])
	})

	t.Run("BinaryFiles", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "match",
		}, func(fs afero.Fs) {
			data := make([]byte, 100)
			data[0] = 0 // null byte
			copy(data[1:], []byte("match me"))
			require.NoError(t, afero.WriteFile(fs, "/root/binary.bin", data, 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 0)
	})

	t.Run("MaxMatches", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "match",
		}, func(fs afero.Fs) {
			var sb strings.Builder
			for i := 0; i < 150; i++ {
				sb.WriteString(fmt.Sprintf("match %d\n", i))
			}
			require.NoError(t, afero.WriteFile(fs, "/root/many.txt", []byte(sb.String()), 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 100)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := setupAndExec(ctx, map[string]interface{}{
			"path":    "/root",
			"pattern": "match",
		}, func(fs afero.Fs) {
			require.NoError(t, fs.MkdirAll("/root", 0755))
			require.NoError(t, afero.WriteFile(fs, "/root/file.txt", []byte("match"), 0644))
		})
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("FileSizeLimit", func(t *testing.T) {
		res, err := setupAndExec(context.Background(), map[string]interface{}{
			"path":    "/root",
			"pattern": "match",
		}, func(fs afero.Fs) {
			size := 10*1024*1024 + 1
			data := make([]byte, size)
			copy(data[size-5:], []byte("match"))
			require.NoError(t, afero.WriteFile(fs, "/root/large.bin", data, 0644))
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 0)
	})
}
