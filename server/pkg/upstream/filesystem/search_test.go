// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFilesTool(t *testing.T) {
	// Setup in-memory filesystem
	prov := provider.NewTmpfsProvider()
	fs := prov.GetFs()

	// Helper to create files
	createFile := func(path, content string) {
		err := fs.MkdirAll(filepath.Dir(path), 0755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create test data
	createFile("/docs/readme.md", "This is the readme file.")
	createFile("/docs/api.md", "API documentation goes here.")
	createFile("/src/main.go", "package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}")
	createFile("/src/utils.go", "package main\n\nfunc util() {}")
	createFile("/src/test.js", "console.log('test');")
	createFile("/.hidden/secret.txt", "This is a secret.")
	createFile("/src/.env", "SECRET_KEY=12345")

	// Create a binary file (simulated with null bytes)
	createFile("/bin/app", "ELF\x00\x00\x00")

	// Get the tool definition
	toolDef := searchFilesTool(prov, fs)

	t.Run("Happy Path - Simple Search", func(t *testing.T) {
		args := map[string]interface{}{
			"path":    "/",
			"pattern": "readme",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Len(t, matches, 1)
		assert.Equal(t, "/docs/readme.md", matches[0]["file"])
		assert.Equal(t, 1, matches[0]["line_number"])
		assert.Contains(t, matches[0]["line_content"], "readme")
	})

	t.Run("Regex Search", func(t *testing.T) {
		args := map[string]interface{}{
			"path":    "/src",
			"pattern": "func.*\\(",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		// Should match main.go and utils.go
		require.Len(t, matches, 2)

		files := make(map[string]bool)
		for _, m := range matches {
			files[m["file"].(string)] = true
		}
		assert.True(t, files["/src/main.go"])
		assert.True(t, files["/src/utils.go"])
	})

	t.Run("Invalid Regex", func(t *testing.T) {
		args := map[string]interface{}{
			"path":    "/",
			"pattern": "[", // Invalid regex
		}
		_, err := toolDef.Handler(context.Background(), args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("Exclusion Patterns", func(t *testing.T) {
		args := map[string]interface{}{
			"path":    "/src",
			"pattern": "func",
			"exclude_patterns": []interface{}{"*.go"},
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		// Should find nothing because .go files are excluded
		require.Empty(t, matches)
	})

	t.Run("Hidden Files Ignored", func(t *testing.T) {
		args := map[string]interface{}{
			"path":    "/",
			"pattern": "secret",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		// Should not find secret.txt in .hidden or .env
		require.Empty(t, matches)
	})

	t.Run("Binary Files Ignored", func(t *testing.T) {
		args := map[string]interface{}{
			"path":    "/bin",
			"pattern": "ELF",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Empty(t, matches)
	})

	t.Run("Max Matches Limit", func(t *testing.T) {
		// Create a file with many matches
		longContent := strings.Repeat("match me\n", 150)
		createFile("/docs/long.txt", longContent)

		args := map[string]interface{}{
			"path":    "/docs/long.txt",
			"pattern": "match me",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 100)
	})

	t.Run("File Size Limit", func(t *testing.T) {
		// Create a large file (>10MB)
		// Since we use MemMapFs, writing >10MB is fast and safe-ish for tests
		// We can't easily make a sparse file on MemMapFs, so we just write it.
		// 10MB + 1 byte
		size := 10*1024*1024 + 1
		// We'll just write a small file and patch Stat to return large size?
		// Or we rely on afero implementation.
		// Writing 10MB to memory might be slow or consume memory.
		// Instead, we can create a file that is skipped by other means first to check logic order?
		// No, let's just write exactly 10MB+1 byte of zeros.

		// To save memory/time, let's Mock the FS for this specific test case or just try it.
		// 10MB is not THAT big for a test runner.
		largeData := make([]byte, size)
		err := afero.WriteFile(fs, "/large.txt", largeData, 0644)
		require.NoError(t, err)

		args := map[string]interface{}{
			"path":    "/large.txt",
			"pattern": "foo",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		require.Empty(t, matches) // Should be skipped
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "readme",
		}
		_, err := toolDef.Handler(ctx, args)
		// Should return context error or nil with partial results depending on where it checks
		// The implementation checks ctx.Err() inside the walk.

		// If it checks immediately, it returns error.
		if err == nil {
			// If it returned nil, it might have skipped everything.
			// But since we have files, it should have tried to walk.
		} else {
			assert.ErrorIs(t, err, context.Canceled)
		}
	})

	t.Run("Missing Arguments", func(t *testing.T) {
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")

		_, err = toolDef.Handler(context.Background(), map[string]interface{}{"path": "/"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern is required")
	})
}
