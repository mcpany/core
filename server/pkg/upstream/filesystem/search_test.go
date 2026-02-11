// Copyright 2026 Author(s) of MCP Any
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

// Mock provider implementation for testing
type mockProvider struct {
	fs afero.Fs
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockProvider) ResolvePath(virtualPath string) (string, error) {
	// Simple passthrough for testing
	return virtualPath, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFilesTool_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Setup files
	require.NoError(t, afero.WriteFile(fs, "/file1.txt", []byte("hello world\nthis is a test\n"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/subdir/file2.txt", []byte("another hello here\n"), 0644))
	require.NoError(t, fs.MkdirAll("/subdir", 0755))

	tool := searchFilesTool(prov, fs)

	// Test search for "hello"
	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "hello",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	require.Len(t, matches, 2)

	// Validate matches
	files := make(map[string]bool)
	for _, m := range matches {
		files[m["file"].(string)] = true
		assert.Contains(t, m["line_content"].(string), "hello")
	}
	assert.True(t, files["/file1.txt"])
	assert.True(t, files["/subdir/file2.txt"])
}

func TestSearchFilesTool_NoMatch(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	require.NoError(t, afero.WriteFile(fs, "/file1.txt", []byte("hello world\n"), 0644))

	tool := searchFilesTool(prov, fs)

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "missing",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Empty(t, matches)
}

func TestSearchFilesTool_InvalidRegex(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	tool := searchFilesTool(prov, fs)

	_, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "[", // Invalid regex
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFilesTool_ExcludePatterns(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	require.NoError(t, afero.WriteFile(fs, "/included.txt", []byte("match me\n"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/excluded.test.js", []byte("match me\n"), 0644))

	tool := searchFilesTool(prov, fs)

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
		"exclude_patterns": []interface{}{"*.test.js"},
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	require.Len(t, matches, 1)
	assert.Equal(t, "/included.txt", matches[0]["file"])
}

func TestSearchFilesTool_HiddenFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	require.NoError(t, afero.WriteFile(fs, "/visible.txt", []byte("match me\n"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/.hidden", []byte("match me\n"), 0644))
	require.NoError(t, fs.MkdirAll("/.git", 0755))
	require.NoError(t, afero.WriteFile(fs, "/.git/config", []byte("match me\n"), 0644))

	tool := searchFilesTool(prov, fs)

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	require.Len(t, matches, 2)

	files := make(map[string]bool)
	for _, m := range matches {
		files[m["file"].(string)] = true
	}
	assert.True(t, files["/visible.txt"])
	assert.True(t, files["/.hidden"])
	assert.False(t, files["/.git/config"]) // Hidden directory content should be skipped
}

func TestSearchFilesTool_BinaryFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	require.NoError(t, afero.WriteFile(fs, "/text.txt", []byte("match me\n"), 0644))
	// Create a "binary" file with null bytes
	binaryContent := append([]byte("match me"), 0x00, 0x00)
	require.NoError(t, afero.WriteFile(fs, "/binary.bin", binaryContent, 0644))

	tool := searchFilesTool(prov, fs)

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	require.Len(t, matches, 1)
	assert.Equal(t, "/text.txt", matches[0]["file"])
}

func TestSearchFilesTool_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create a file with 101 matching lines
	var content strings.Builder
	for i := 0; i < 150; i++ {
		content.WriteString(fmt.Sprintf("match me %d\n", i))
	}
	require.NoError(t, afero.WriteFile(fs, "/many_matches.txt", []byte(content.String()), 0644))

	tool := searchFilesTool(prov, fs)

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100)
}

func TestSearchFilesTool_ContextCancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	require.NoError(t, afero.WriteFile(fs, "/file.txt", []byte("match me\n"), 0644))

	tool := searchFilesTool(prov, fs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := tool.Handler(ctx, map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestSearchFilesTool_MissingArgs(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// Missing path
	_, err := tool.Handler(context.Background(), map[string]interface{}{
		"pattern": "foo",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")

	// Missing pattern
	_, err = tool.Handler(context.Background(), map[string]interface{}{
		"path": "/",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pattern is required")
}
