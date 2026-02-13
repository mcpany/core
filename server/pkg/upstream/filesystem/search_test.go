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

// mockProvider implements provider.Provider for testing purposes
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
	afero.WriteFile(fs, "/foo.txt", []byte("hello world"), 0644)
	afero.WriteFile(fs, "/bar.txt", []byte("goodbye world"), 0644)
	afero.WriteFile(fs, "/baz.txt", []byte("hello universe"), 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "hello",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 2)

	// Verify contents
	found := make(map[string]bool)
	for _, m := range matches {
		file := m["file"].(string)
		found[file] = true
		assert.Contains(t, m["line_content"], "hello")
	}
	assert.True(t, found["/foo.txt"])
	assert.True(t, found["/baz.txt"])
}

func TestSearchFilesTool_NoMatch(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/foo.txt", []byte("foo"), 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "bar",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 0)
}

func TestSearchFilesTool_InvalidRegex(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "(", // Invalid regex
	}

	_, err := tool.Handler(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFilesTool_ExcludePatterns(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/foo.js", []byte("console.log('hello')"), 0644)
	afero.WriteFile(fs, "/foo.test.js", []byte("console.log('hello')"), 0644)
	afero.WriteFile(fs, "/bar.js", []byte("console.log('hello')"), 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "hello",
		"exclude_patterns": []interface{}{
			"*.test.js",
			"bar.js",
		},
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/foo.js", matches[0]["file"])
}

func TestSearchFilesTool_HiddenDirectories(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/visible.txt", []byte("secret"), 0644)

	// Create hidden directory and file inside
	fs.MkdirAll("/.hidden", 0755)
	afero.WriteFile(fs, "/.hidden/secret.txt", []byte("secret"), 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "secret",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/visible.txt", matches[0]["file"])
}

func TestSearchFilesTool_BinaryFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/text.txt", []byte("hello world"), 0644)

	// Create a "binary" file with null bytes
	binaryContent := []byte("hello \x00 world")
	afero.WriteFile(fs, "/binary.bin", binaryContent, 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "hello",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/text.txt", matches[0]["file"])
}

func TestSearchFilesTool_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a file with 150 matching lines
	var sb strings.Builder
	for i := 0; i < 150; i++ {
		sb.WriteString(fmt.Sprintf("match %d\n", i))
	}
	afero.WriteFile(fs, "/many.txt", []byte(sb.String()), 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "match",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100) // Expect capped at 100
}

func TestSearchFilesTool_ContextCancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a large file to ensure we spend some time processing
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString(fmt.Sprintf("line %d\n", i))
	}
	afero.WriteFile(fs, "/large.txt", []byte(sb.String()), 0644)

	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "line",
	}

	_, err := tool.Handler(ctx, args)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
