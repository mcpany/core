// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestSearchFilesTool_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	// Setup files
	require.NoError(t, fs.MkdirAll("/data", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/file1.txt", []byte("hello world"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/data/file2.txt", []byte("hello there"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/data/file3.txt", []byte("goodbye world"), 0644))

	// Execute search
	result, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/data",
		"pattern": "hello",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Len(t, matches, 2)

	// Verify match contents
	files := make(map[string]string)
	for _, m := range matches {
		files[m["file"].(string)] = m["line_content"].(string)
	}

	assert.Contains(t, files, "/data/file1.txt")
	assert.Contains(t, files, "/data/file2.txt")
	assert.Equal(t, "hello world", files["/data/file1.txt"])
	assert.Equal(t, "hello there", files["/data/file2.txt"])
}

func TestSearchFilesTool_NoMatch(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	require.NoError(t, fs.MkdirAll("/data", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/file1.txt", []byte("hello world"), 0644))

	result, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/data",
		"pattern": "foobar",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Empty(t, matches)
}

func TestSearchFilesTool_InvalidRegex(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	_, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/data",
		"pattern": "[", // Invalid regex
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFilesTool_ExcludePatterns(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	require.NoError(t, fs.MkdirAll("/data", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/file1.txt", []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/data/test.js", []byte("match me"), 0644)) // Should be excluded

	result, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":             "/data",
		"pattern":          "match",
		"exclude_patterns": []interface{}{"*.js"},
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/data/file1.txt", matches[0]["file"])
}

func TestSearchFilesTool_HiddenFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	require.NoError(t, fs.MkdirAll("/data/.hidden", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/.hidden/file.txt", []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/data/visible.txt", []byte("match me"), 0644))

	result, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/data",
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/data/visible.txt", matches[0]["file"])
}

func TestSearchFilesTool_BinaryFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	require.NoError(t, fs.MkdirAll("/data", 0755))

	// Create a binary file (contains null byte)
	binaryContent := []byte("match me")
	binaryContent = append(binaryContent, 0x00)
	require.NoError(t, afero.WriteFile(fs, "/data/binary.bin", binaryContent, 0644))

	require.NoError(t, afero.WriteFile(fs, "/data/text.txt", []byte("match me"), 0644))

	result, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/data",
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/data/text.txt", matches[0]["file"])
}

func TestSearchFilesTool_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	require.NoError(t, fs.MkdirAll("/data", 0755))

	// Create more than 100 matching lines in a single file
	var content string
	for i := 0; i < 150; i++ {
		content += fmt.Sprintf("match line %d\n", i)
	}
	require.NoError(t, afero.WriteFile(fs, "/data/many_matches.txt", []byte(content), 0644))

	result, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/data",
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := result["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100)
}

func TestSearchFilesTool_ContextCancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	require.NoError(t, fs.MkdirAll("/data", 0755))

	// Create many files to ensure the walk takes some time/iterations
	for i := 0; i < 100; i++ {
		require.NoError(t, afero.WriteFile(fs, fmt.Sprintf("/data/file%d.txt", i), []byte("match me"), 0644))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := toolDef.Handler(ctx, map[string]interface{}{
		"path":    "/data",
		"pattern": "match",
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
