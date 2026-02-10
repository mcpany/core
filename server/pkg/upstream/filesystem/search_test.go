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

// mockProvider implements provider.Provider for testing
type mockProvider struct {
	fs afero.Fs
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockProvider) ResolvePath(virtualPath string) (string, error) {
	// For testing, we assume virtual path maps directly to the mock FS path
	return virtualPath, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFilesTool_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Setup files
	err := afero.WriteFile(fs, "/foo.txt", []byte("hello world\nthis is a test\nend of file"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/bar.txt", []byte("another file\nwith some content\nhello there"), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)

	// Search for "hello"
	args := map[string]interface{}{
		"path":    "/",
		"pattern": "hello",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
	require.NotNil(t, res)

	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, matches, 2)

	// Validate matches
	files := make(map[string]bool)
	for _, m := range matches {
		files[m["file"].(string)] = true
		assert.Contains(t, m["line_content"], "hello")
	}
	assert.True(t, files["/foo.txt"])
	assert.True(t, files["/bar.txt"])
}

func TestSearchFilesTool_NoMatch(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	err := afero.WriteFile(fs, "/foo.txt", []byte("hello world"), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "notfound",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, matches, 0)
}

func TestSearchFilesTool_InvalidRegex(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "(invalid regex",
	}

	_, err := tool.Handler(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFilesTool_ExcludePatterns(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	err := afero.WriteFile(fs, "/include.txt", []byte("target"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/exclude.test.js", []byte("target"), 0644)
	require.NoError(t, err)

	err = fs.Mkdir("/node_modules", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/node_modules/foo.js", []byte("target"), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "target",
		"exclude_patterns": []interface{}{"*.test.js", "node_modules"},
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)

	// Expect only include.txt
	// exclude.test.js matches *.test.js
	// node_modules/foo.js matches node_modules (directory exclusion logic)

	foundFiles := make(map[string]bool)
	for _, m := range matches {
		foundFiles[m["file"].(string)] = true
	}

	assert.True(t, foundFiles["/include.txt"], "include.txt should be found")
	assert.False(t, foundFiles["/exclude.test.js"], "exclude.test.js should be excluded")
	assert.False(t, foundFiles["/node_modules/foo.js"], "node_modules should be excluded")
}

func TestSearchFilesTool_HiddenFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	err := afero.WriteFile(fs, "/.hidden", []byte("target"), 0644)
	require.NoError(t, err)
	err = fs.Mkdir("/.git", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/.git/config", []byte("target"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/visible.txt", []byte("target"), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "target",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)

	foundFiles := make(map[string]bool)
	for _, m := range matches {
		foundFiles[m["file"].(string)] = true
	}

	assert.True(t, foundFiles["/visible.txt"])
	assert.False(t, foundFiles["/.git/config"])
	// Hidden files are NOT skipped by default logic, only directories
	assert.True(t, foundFiles["/.hidden"])
}

func TestSearchFilesTool_BinaryFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create a file with binary content.
	// http.DetectContentType will detect octet-stream for binary data.
	binaryContent := make([]byte, 100)
	for i := 0; i < 100; i++ {
		binaryContent[i] = byte(i)
	}
	// Add the pattern to verify it's skipped despite matching regex
	copy(binaryContent[60:], []byte("target"))

	err := afero.WriteFile(fs, "/binary.bin", binaryContent, 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/text.txt", []byte("target"), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "target",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, matches, 1)
	assert.Equal(t, "/text.txt", matches[0]["file"])
}

func TestSearchFilesTool_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create a file with many matches
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		sb.WriteString("match me\n")
	}
	err := afero.WriteFile(fs, "/many.txt", []byte(sb.String()), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	}

	res, err := tool.Handler(context.Background(), args)
	require.NoError(t, err)
	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)

	// The code has maxMatches = 100
	assert.Len(t, matches, 100)
}

func TestSearchFilesTool_ContextCancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create many files to ensure the walk takes some time/iterations
	for i := 0; i < 100; i++ {
		err := afero.WriteFile(fs, fmt.Sprintf("/file%d.txt", i), []byte("target"), 0644)
		require.NoError(t, err)
	}

	tool := searchFilesTool(prov, fs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	args := map[string]interface{}{
		"path":    "/",
		"pattern": "target",
	}

	_, err := tool.Handler(ctx, args)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
