// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements provider.Provider for testing
type mockProvider struct {
	rootPath string
	fs       afero.Fs
}

// Ensure mockProvider implements provider.Provider
var _ provider.Provider = (*mockProvider)(nil)

func (m *mockProvider) ResolvePath(path string) (string, error) {
	// Simple resolution logic for testing
	if filepath.IsAbs(path) {
		if !strings.HasPrefix(path, m.rootPath) {
			return "", fmt.Errorf("path outside allowed root")
		}
		return path, nil
	}
	return filepath.Join(m.rootPath, path), nil
}

func (m *mockProvider) Close() error {
	return nil
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func TestSearchFiles(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	prov := &mockProvider{
		rootPath: "/data",
		fs:       fs,
	}
	searchTool := searchFilesTool(prov, fs)

	// Helper to create files
	createFile := func(path string, content string) {
		dir := filepath.Dir(path)
		_ = fs.MkdirAll(dir, 0755)
		_ = afero.WriteFile(fs, path, []byte(content), 0644)
	}

	// Create test files
	createFile("/data/file1.txt", "hello world")
	createFile("/data/file2.txt", "foo bar baz")
	createFile("/data/subdir/file3.txt", "hello universe")
	createFile("/data/code.js", "console.log('hello');")
	createFile("/data/test.js", "test('hello');")
	createFile("/data/.hidden", "hello secret")
	createFile("/data/binary.bin", string([]byte{0, 1, 2, 3, 4, 5})+"hello binary")
	createFile("/data/large.txt", strings.Repeat("a", 10*1024*1024+1)+"hello large")

	t.Run("Happy Path - Simple String", func(t *testing.T) {
		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "hello",
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		// Expected matches: file1.txt, subdir/file3.txt, code.js, test.js
		// .hidden should be skipped
		// binary.bin should be skipped
		// large.txt should be skipped
		assert.Len(t, matches, 4)

		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
			assert.Contains(t, m["line_content"], "hello")
		}
		assert.True(t, foundFiles["/data/file1.txt"])
		assert.True(t, foundFiles["/data/subdir/file3.txt"])
		assert.True(t, foundFiles["/data/code.js"])
		assert.True(t, foundFiles["/data/test.js"])
	})

	t.Run("Regex Support", func(t *testing.T) {
		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "foo.*baz",
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/data/file2.txt", matches[0]["file"])
		assert.Equal(t, "foo bar baz", matches[0]["line_content"])
	})

	t.Run("Invalid Regex", func(t *testing.T) {
		_, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "[", // Invalid regex
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("Exclusion Patterns", func(t *testing.T) {
		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":             "/data",
			"pattern":          "hello",
			"exclude_patterns": []interface{}{"*.js"},
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		// Should exclude code.js and test.js
		assert.Len(t, matches, 2)
		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}
		assert.True(t, foundFiles["/data/file1.txt"])
		assert.True(t, foundFiles["/data/subdir/file3.txt"])
		assert.False(t, foundFiles["/data/code.js"])
	})

	t.Run("Hidden Files Skipped", func(t *testing.T) {
		// create a hidden file that matches
		createFile("/data/.secret", "hello secret")

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "secret",
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 0)
	})

	t.Run("Binary Files Skipped", func(t *testing.T) {
		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "binary",
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 0)
	})

	t.Run("Max Matches Limit", func(t *testing.T) {
		// Create a file with many matches
		manyMatches := strings.Repeat("match\n", 150)
		createFile("/data/many.txt", manyMatches)

		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data/many.txt",
			"pattern": "match",
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 100) // Capped at 100
	})

	t.Run("File Size Limit", func(t *testing.T) {
		// heavily padded file created in Setup
		res, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "large",
		})
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		// Should not find "hello large" in large.txt
		for _, m := range matches {
			assert.NotEqual(t, "/data/large.txt", m["file"])
		}
	})

	t.Run("Missing Arguments", func(t *testing.T) {
		_, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"pattern": "hello",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")

		_, err = searchTool.Handler(context.Background(), map[string]interface{}{
			"path": "/data",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pattern is required")
	})

	t.Run("Resolve Path Error", func(t *testing.T) {
		_, err := searchTool.Handler(context.Background(), map[string]interface{}{
			"path":    "/outside/root",
			"pattern": "hello",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path outside allowed root")
	})
}
