// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSearchTest creates a TmpfsProvider and returns the tool function for testing
func setupSearchTest(t *testing.T) (func(context.Context, map[string]interface{}) (map[string]interface{}, error), afero.Fs) {
	prov := provider.NewTmpfsProvider()
	fs := prov.GetFs()
	toolDef := searchFilesTool(prov, fs)
	return toolDef.Handler, fs
}

func TestSearchFiles_HappyPath(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create a file to search
	err := afero.WriteFile(fs, "/test.txt", []byte("hello world\nfind me here\nanother line"), 0644)
	require.NoError(t, err)

	// Search
	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "find me",
	})
	require.NoError(t, err)

	resMap, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, resMap, 1)

	match := resMap[0]
	assert.Equal(t, "/test.txt", match["file"])
	assert.Equal(t, 2, match["line_number"])
	assert.Equal(t, "find me here", match["line_content"])
}

func TestSearchFiles_InputValidation(t *testing.T) {
	handler, _ := setupSearchTest(t)

	// Missing path
	_, err := handler(context.Background(), map[string]interface{}{
		"pattern": "foo",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")

	// Missing pattern
	_, err = handler(context.Background(), map[string]interface{}{
		"path": "/",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pattern is required")
}

func TestSearchFiles_RegexValidation(t *testing.T) {
	handler, _ := setupSearchTest(t)

	// Invalid regex
	_, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "[invalid",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFiles_Exclusions(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create files
	require.NoError(t, afero.WriteFile(fs, "/include.txt", []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/exclude.log", []byte("match me"), 0644))
	require.NoError(t, fs.MkdirAll("/node_modules", 0755))
	require.NoError(t, afero.WriteFile(fs, "/node_modules/foo.js", []byte("match me"), 0644))

	// Search
	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
		"exclude_patterns": []interface{}{
			"*.log",
			"node_modules",
		},
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should match include.txt only
	assert.Len(t, matches, 1)
	assert.Equal(t, "/include.txt", matches[0]["file"])
}

func TestSearchFiles_HiddenFiles(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create hidden file and directory
	require.NoError(t, afero.WriteFile(fs, "/.hidden", []byte("match me"), 0644))
	require.NoError(t, fs.MkdirAll("/.git", 0755))
	require.NoError(t, afero.WriteFile(fs, "/.git/config", []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/visible.txt", []byte("match me"), 0644))

	// Search
	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should match visible.txt only
	assert.Len(t, matches, 1)
	assert.Equal(t, "/visible.txt", matches[0]["file"])
}

func TestSearchFiles_BinaryFiles(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create binary file (with null byte)
	require.NoError(t, afero.WriteFile(fs, "/binary.bin", []byte("match me\x00binary"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/text.txt", []byte("match me text"), 0644))

	// Search
	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should match text.txt only
	assert.Len(t, matches, 1)
	assert.Equal(t, "/text.txt", matches[0]["file"])
}

func TestSearchFiles_LargeFiles(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create large file (> 10MB)
	// 10MB + 1 byte
	size := 10*1024*1024 + 1
	// Create a sparse file or just write zeros? MemMapFs might actually allocate.
	// 10MB is fine.
	data := make([]byte, size)
	// Put "match me" at the beginning so if it wasn't skipped it would be found
	copy(data, []byte("match me"))

	require.NoError(t, afero.WriteFile(fs, "/large.txt", data, 0644))
	require.NoError(t, afero.WriteFile(fs, "/normal.txt", []byte("match me"), 0644))

	// Search
	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should match normal.txt only
	assert.Len(t, matches, 1)
	assert.Equal(t, "/normal.txt", matches[0]["file"])
}

func TestSearchFiles_MaxMatches(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create a file with > 100 matches
	var sb strings.Builder
	for i := 0; i < 150; i++ {
		sb.WriteString("match me\n")
	}
	require.NoError(t, afero.WriteFile(fs, "/many_matches.txt", []byte(sb.String()), 0644))

	// Search
	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should be capped at 100
	assert.Len(t, matches, 100)
}

func TestSearchFiles_ContextCancellation(t *testing.T) {
	handler, fs := setupSearchTest(t)

	// Create enough content to make search take some time (mock fs is fast though)
	// We might not be able to catch it mid-flight easily with MemMapFs,
	// but we can test that passing an already cancelled context returns error.
	require.NoError(t, afero.WriteFile(fs, "/test.txt", []byte("match me"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := handler(ctx, map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
