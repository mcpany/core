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

// Mock provider for testing
type mockProvider struct {
	fs afero.Fs
}

func (m *mockProvider) ResolvePath(path string) (string, error) {
	return path, nil
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFiles_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(filepath.Join(rootDir, "subdir"), 0755))

	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "file1.txt"), []byte("hello world"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "subdir", "file2.txt"), []byte("foo bar hello universe"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "file3.txt"), []byte("goodbye"), 0644))

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "hello",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 2)

	// Verify content of matches
	paths := []string{}
	for _, m := range matches {
		paths = append(paths, m["file"].(string))
		assert.Contains(t, m["line_content"], "hello")
	}
	assert.Contains(t, paths, filepath.Join(rootDir, "file1.txt"))
	assert.Contains(t, paths, filepath.Join(rootDir, "subdir", "file2.txt"))
}

func TestSearchFiles_Regex(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(rootDir, 0755))

	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "test.log"), []byte("Error: code 500\nInfo: all good"), 0644))

	// Case insensitive search using regex flag
	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "(?i)error",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, filepath.Join(rootDir, "test.log"), matches[0]["file"])
	assert.Equal(t, "Error: code 500", matches[0]["line_content"])
}

func TestSearchFiles_InvalidRegex(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	_, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "[", // Invalid regex
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFiles_Exclusion(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(rootDir, 0755))

	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "keep.txt"), []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "skip.log"), []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "ignore.me"), []byte("match me"), 0644))

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "match",
		"exclude_patterns": []interface{}{"*.log", "ignore.*"},
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, filepath.Join(rootDir, "keep.txt"), matches[0]["file"])
}

func TestSearchFiles_Hidden(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(filepath.Join(rootDir, ".git"), 0755))

	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, ".hidden"), []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, ".git", "config"), []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "visible.txt"), []byte("match me"), 0644))

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	// Current implementation only skips hidden directories, not hidden files.
	// So we expect 2 matches: visible.txt and .hidden, but NOT .git/config
	assert.Len(t, matches, 2)

	paths := []string{}
	for _, m := range matches {
		paths = append(paths, m["file"].(string))
	}
	assert.Contains(t, paths, filepath.Join(rootDir, "visible.txt"))
	assert.Contains(t, paths, filepath.Join(rootDir, ".hidden"))
	assert.NotContains(t, paths, filepath.Join(rootDir, ".git", "config"))
}

func TestSearchFiles_Binary(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(rootDir, 0755))

	// Create binary file (null bytes)
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03}
	// Append "match" to binary file to ensure it would match if treated as text
	binaryContent = append(binaryContent, []byte("match")...)

	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "binary.bin"), binaryContent, 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "text.txt"), []byte("match"), 0644))

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, filepath.Join(rootDir, "text.txt"), matches[0]["file"])
}

func TestSearchFiles_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(rootDir, 0755))

	// Create 105 matches
	for i := 0; i < 105; i++ {
		fname := fmt.Sprintf("file_%d.txt", i)
		require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, fname), []byte("match me"), 0644))
	}

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100)
}

func TestSearchFiles_FileSizeLimit(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(rootDir, 0755))

	// 10MB + 1 byte
	largeSize := 10*1024*1024 + 1
	// Use sparse file if filesystem supports it, or just write zeroes.
	// For MemMapFs, we have to write it. It might be slow but it's memory.
	// We'll write a small buffer repeatedly.

	// Actually, let's just make it slightly larger than limit.
	// We need to write "match" in it to potentially match.
	// But simply checking size is enough to skip.

	// Create a dummy file with large size without allocating full memory if possible?
	// MemMapFs keeps it in memory. 10MB is fine for test environment usually.

	largeContent := make([]byte, largeSize)
	// Put "match" at the beginning
	copy(largeContent, []byte("match"))

	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "large.txt"), largeContent, 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "normal.txt"), []byte("match"), 0644))

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    rootDir,
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, filepath.Join(rootDir, "normal.txt"), matches[0]["file"])
}

func TestSearchFiles_Cancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	rootDir := "/test_root"
	require.NoError(t, fs.MkdirAll(rootDir, 0755))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(rootDir, "file.txt"), []byte("match"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := toolDef.Handler(ctx, map[string]interface{}{
		"path":    rootDir,
		"pattern": "match",
	})

	// Should return error or empty matches depending on where it checks.
	// The code checks `ctx.Err()` inside the walk loop.
	// Depending on afero implementation, Walk might return error immediately if ctx is done?
	// afero.Walk doesn't take context. The callback checks context.
	// So it should return ctx.Err()

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestSearchFiles_InputValidation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)

	// Missing path
	_, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"pattern": "match",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")

	// Missing pattern
	_, err = toolDef.Handler(context.Background(), map[string]interface{}{
		"path": "/",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pattern is required")
}
