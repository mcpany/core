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
	// Simple mock implementation: return virtualPath as is, or error if restricted.
	// For testing, we assume all paths are allowed unless explicitly disallowed.
	if strings.Contains(virtualPath, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	return virtualPath, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFilesTool(t *testing.T) {
	// Setup filesystem
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Helper function to create files
	createFile := func(path, content string) {
		err := afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create test files
	createFile("root.txt", "hello world\nthis is a test\n")
	createFile("other.txt", "foo bar\nbaz qux\n")

	err := fs.Mkdir("subdir", 0755)
	require.NoError(t, err)
	createFile("subdir/sub.txt", "hello form subdir\nanother line\n")

	createFile("exclude.log", "this should be excluded\n")
	createFile(".hidden", "this is hidden\n")

	// Create binary file (null byte)
	createFile("binary.bin", "binary\x00content\n")

	// Create a large file (simulated with content)
	// MemMapFs keeps everything in memory, so we can't make it *too* big without OOM risk in test,
	// but 10MB is manageable. The search tool limit is > 10*1024*1024.
	// We'll skip creating a real 10MB file for now to keep test fast, but we can test logic if we could mock file info size.
	// Since we can't easily mock FileInfo size with MemMapFs without patching afero, we will skip the large file test case
	// or create a small file and just trust the logic (which is standard `info.Size() > limit`).
	// Alternatively, we can create a sparse file if the filesystem supported it, but MemMapFs doesn't.
	// Let's create a file slightly larger than 10MB to test the limit. 10MB + 1 byte.
	// sparseFileContent := make([]byte, 10*1024*1024+1)
	// err = afero.WriteFile(fs, "large.txt", sparseFileContent, 0644)
	// require.NoError(t, err)
	// NOTE: Writing 10MB in memory for a test might be slow/heavy. Let's rely on unit logic inspection for now,
	// or try a smaller limit if we could configure it. The limit is hardcoded.
	// We will skip the large file test to avoid slowing down the test suite, unless necessary.

	tests := []struct {
		name          string
		args          map[string]interface{}
		expectedMatch int
		wantErr       bool
		validate      func(t *testing.T, matches []map[string]interface{})
	}{
		{
			name: "Basic Search",
			args: map[string]interface{}{
				"path":    ".",
				"pattern": "hello",
			},
			expectedMatch: 2, // root.txt and subdir/sub.txt
			validate: func(t *testing.T, matches []map[string]interface{}) {
				files := make(map[string]bool)
				for _, m := range matches {
					files[m["file"].(string)] = true
				}
				assert.True(t, files["root.txt"])
				assert.True(t, files["subdir/sub.txt"])
			},
		},
		{
			name: "Subdirectory Search",
			args: map[string]interface{}{
				"path":    "subdir",
				"pattern": "hello",
			},
			expectedMatch: 1, // subdir/sub.txt
			validate: func(t *testing.T, matches []map[string]interface{}) {
				match := matches[0]
				// The tool joins input path with relative path.
				// Input: "subdir", Rel: "sub.txt" -> "subdir/sub.txt"
				assert.Equal(t, "subdir/sub.txt", match["file"])
			},
		},
		{
			name: "No Match",
			args: map[string]interface{}{
				"path":    ".",
				"pattern": "nonexistent",
			},
			expectedMatch: 0,
		},
		{
			name: "Exclude Pattern",
			args: map[string]interface{}{
				"path":             ".",
				"pattern":          "excluded",
				"exclude_patterns": []interface{}{"*.log"},
			},
			expectedMatch: 0, // exclude.log has "excluded" but should be skipped
		},
		{
			name: "Binary File Skip",
			args: map[string]interface{}{
				"path":    ".",
				"pattern": "binary",
			},
			expectedMatch: 0, // binary.bin has "binary" but contains null byte
		},
		{
			name: "Hidden File Skip",
			args: map[string]interface{}{
				"path":    ".",
				"pattern": "hidden",
			},
			expectedMatch: 0, // .hidden should be skipped
		},
		{
			name: "Invalid Regex",
			args: map[string]interface{}{
				"path":    ".",
				"pattern": "[a-",
			},
			wantErr: true,
		},
		{
			name: "Missing Path",
			args: map[string]interface{}{
				"pattern": "hello",
			},
			wantErr: true,
		},
		{
			name: "Missing Pattern",
			args: map[string]interface{}{
				"path": ".",
			},
			wantErr: true,
		},
		{
			name: "Resolve Error",
			args: map[string]interface{}{
				"path":    "../illegal",
				"pattern": "foo",
			},
			wantErr: true, // Mock provider returns error for ".."
		},
	}

	tool := searchFilesTool(prov, fs)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tool.Handler(context.Background(), tc.args)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			matches := res["matches"].([]map[string]interface{})
			assert.Len(t, matches, tc.expectedMatch)

			if tc.validate != nil {
				tc.validate(t, matches)
			}
		})
	}
}

func TestSearchFiles_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create a file with 150 matches
	var content strings.Builder
	for i := 0; i < 150; i++ {
		content.WriteString("match me\n")
	}
	err := afero.WriteFile(fs, "many.txt", []byte(content.String()), 0644)
	require.NoError(t, err)

	tool := searchFilesTool(prov, fs)
	res, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    ".",
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100, "Should be capped at 100 matches")
}

func TestSearchFiles_ContextCancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}

	// Create many files to ensure search takes some time
	for i := 0; i < 1000; i++ {
		afero.WriteFile(fs, fmt.Sprintf("file_%d.txt", i), []byte("match\n"), 0644)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tool := searchFilesTool(prov, fs)
	_, err := tool.Handler(ctx, map[string]interface{}{
		"path":    ".",
		"pattern": "match",
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
