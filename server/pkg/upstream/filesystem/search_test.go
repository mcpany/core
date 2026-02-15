// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"path/filepath"
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

func (m *mockProvider) Close() error { return nil }
func (m *mockProvider) GetFs() afero.Fs { return m.fs }
func (m *mockProvider) ResolvePath(virtualPath string) (string, error) {
	return virtualPath, nil
}

func TestSearchFilesTool(t *testing.T) {
	setup := func(t *testing.T) (afero.Fs, filesystemToolDef, func(string, string)) {
		fs := afero.NewMemMapFs()
		prov := &mockProvider{fs: fs}
		searchTool := searchFilesTool(prov, fs)

		createFile := func(path, content string) {
			err := fs.MkdirAll(filepath.Dir(path), 0755)
			require.NoError(t, err)
			err = afero.WriteFile(fs, path, []byte(content), 0644)
			require.NoError(t, err)
		}
		return fs, searchTool, createFile
	}

	t.Run("HappyPath", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		createFile("/data/file1.txt", "hello world\nthis is a test\nfind me here")
		createFile("/data/file2.txt", "nothing here")

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "find me",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "/data/file1.txt", matches[0]["file"])
		assert.Equal(t, 3, matches[0]["line_number"])
		assert.Equal(t, "find me here", matches[0]["line_content"])
	})

	t.Run("RegexSupport", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		createFile("/data/regex.txt", "foo123bar")

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "foo\\d+bar",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "foo123bar", matches[0]["line_content"])
	})

	t.Run("InvalidRegex", func(t *testing.T) {
		_, searchTool, _ := setup(t)
		_, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "[invalid",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("ExclusionPatterns", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		createFile("/data/include.txt", "match")
		createFile("/data/exclude.test.js", "match")

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
			"exclude_patterns": []interface{}{"*.test.js"},
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "/data/include.txt", matches[0]["file"])
	})

	t.Run("HiddenFiles", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		createFile("/data/.hidden/file.txt", "match")
		createFile("/data/visible.txt", "match")

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "/data/visible.txt", matches[0]["file"])
	})

	t.Run("BinaryFiles", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		// Create a file with null byte to simulate binary
		createFile("/data/binary.bin", "match\x00binary")
		createFile("/data/text.txt", "match text")

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "/data/text.txt", matches[0]["file"])
	})

	t.Run("MaxMatches", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		// Create a file with many matches
		var content strings.Builder
		for i := 0; i < 150; i++ {
			content.WriteString(fmt.Sprintf("match %d\n", i))
		}
		createFile("/data/many.txt", content.String())

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 100)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		_, searchTool, createFile := setup(t)
		createFile("/data/cancel.txt", "match")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := searchTool.Handler(ctx, map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		})
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("FileSizeLimit", func(t *testing.T) {
		fs, searchTool, createFile := setup(t)
		// Create a large file (>10MB)
		// We can't easily create 10MB in memory without consuming memory,
		// but 10MB + 1 byte is enough.
		size := 10*1024*1024 + 1
		// Use sparse file trick or just write empty bytes?
		// Afero MemMapFs stores content in byte slice, so we have to allocate it.
		// 10MB is manageable in test environment.
		largeContent := make([]byte, size)
		err := afero.WriteFile(fs, "/data/large.txt", largeContent, 0644)
		require.NoError(t, err)

		createFile("/data/small.txt", "match")

		// We need the large file to theoretically match if it wasn't skipped.
		// But since it's all zeros, it won't match "match".
		// However, the code skips BEFORE reading content (based on size).
		// So checking that it doesn't error or crash is good, but verifying it was skipped
		// requires us to know it wasn't read.
		// The test is that we don't get a match even if it had the content.

		// Let's make the large content have "match" at the beginning,
		// so if it WAS read, it WOULD match.
		copy(largeContent, []byte("match"))
		err = afero.WriteFile(fs, "/data/large.txt", largeContent, 0644)
		require.NoError(t, err)

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "/data/small.txt", matches[0]["file"])
	})
}
