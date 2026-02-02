// Copyright 2026 Author(s) of MCP Any
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

// Mock provider for testing
type mockProvider struct {
	fs afero.Fs
}

func (m *mockProvider) GetFs() afero.Fs {
	return m.fs
}

func (m *mockProvider) ResolvePath(virtualPath string) (string, error) {
	// Identity mapping for testing
	return virtualPath, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFilesTool(t *testing.T) {
	t.Parallel()

	setup := func() (afero.Fs, filesystemToolDef) {
		fs := afero.NewMemMapFs()
		prov := &mockProvider{fs: fs}
		toolDef := searchFilesTool(prov, fs)
		return fs, toolDef
	}

	t.Run("Happy Path", func(t *testing.T) {
		fs, toolDef := setup()

		// Create files
		require.NoError(t, afero.WriteFile(fs, "/foo.txt", []byte("hello world\nthis is a test\nworld is big"), 0644))
		require.NoError(t, afero.WriteFile(fs, "/bar.txt", []byte("another file"), 0644))
		require.NoError(t, fs.MkdirAll("/nested", 0755))
		require.NoError(t, afero.WriteFile(fs, "/nested/baz.txt", []byte("hello nested world"), 0644))

		// Search for "world"
		args := map[string]interface{}{
			"path":    "/",
			"pattern": "world",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})

		assert.Len(t, matches, 3)

		// Verify matches
		files := make(map[string]int)
		for _, m := range matches {
			f := m["file"].(string)
			files[f]++
			if f == "/foo.txt" {
				if m["line_number"].(int) == 1 {
					assert.Equal(t, "hello world", m["line_content"])
				} else if m["line_number"].(int) == 3 {
					assert.Equal(t, "world is big", m["line_content"])
				}
			}
			if f == "/nested/baz.txt" {
				assert.Equal(t, 1, m["line_number"])
				assert.Equal(t, "hello nested world", m["line_content"])
			}
		}
		assert.Equal(t, 2, files["/foo.txt"])
		assert.Equal(t, 1, files["/nested/baz.txt"])
	})

	t.Run("No Match", func(t *testing.T) {
		fs, toolDef := setup()
		require.NoError(t, afero.WriteFile(fs, "/foo.txt", []byte("hello"), 0644))

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "world",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		assert.Empty(t, matches)
	})

	t.Run("Invalid Regex", func(t *testing.T) {
		_, toolDef := setup()
		args := map[string]interface{}{
			"path":    "/",
			"pattern": "[", // Invalid regex
		}
		_, err := toolDef.Handler(context.Background(), args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("Exclusions", func(t *testing.T) {
		fs, toolDef := setup()
		require.NoError(t, afero.WriteFile(fs, "/include.txt", []byte("match"), 0644))
		require.NoError(t, afero.WriteFile(fs, "/exclude.log", []byte("match"), 0644))
		require.NoError(t, afero.WriteFile(fs, "/test.js", []byte("match"), 0644))

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "match",
			"exclude_patterns": []interface{}{
				"*.log",
				"*.js",
			},
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		assert.Len(t, matches, 1)
		assert.Equal(t, "/include.txt", matches[0]["file"])
	})

	t.Run("Hidden Files", func(t *testing.T) {
		fs, toolDef := setup()
		// .hidden file should be found (only hidden dirs are skipped)
		require.NoError(t, afero.WriteFile(fs, "/.hidden", []byte("match"), 0644))
		// .git directory should be skipped
		require.NoError(t, fs.MkdirAll("/.git", 0755))
		require.NoError(t, afero.WriteFile(fs, "/.git/config", []byte("match"), 0644))
		require.NoError(t, afero.WriteFile(fs, "/visible", []byte("match"), 0644))

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "match",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		// Should find /visible and /.hidden, but NOT /.git/config
		assert.Len(t, matches, 2)

		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}
		assert.True(t, foundFiles["/visible"])
		assert.True(t, foundFiles["/.hidden"])
		assert.False(t, foundFiles["/.git/config"])
	})

	t.Run("Binary Files", func(t *testing.T) {
		fs, toolDef := setup()
		// Create a file with null bytes
		binaryContent := append([]byte("match"), 0, 0, 0)
		require.NoError(t, afero.WriteFile(fs, "/binary.bin", binaryContent, 0644))
		require.NoError(t, afero.WriteFile(fs, "/text.txt", []byte("match"), 0644))

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "match",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		assert.Len(t, matches, 1)
		assert.Equal(t, "/text.txt", matches[0]["file"])
	})

	t.Run("Size Limit", func(t *testing.T) {
		fs, toolDef := setup()
		// > 10MB
		size := 10*1024*1024 + 1
		// We use a sparse file trick if supported, or just write a large file.
		// MemMapFs keeps it in memory, so 10MB is fine for testing.
		// Writing 10MB of 'a's then 'match'

		f, err := fs.Create("/large.txt")
		require.NoError(t, err)
		// Seek to end and write
		_, err = f.Seek(int64(size-5), 0)
		require.NoError(t, err)
		_, err = f.Write([]byte("match"))
		require.NoError(t, err)
		f.Close()

		require.NoError(t, afero.WriteFile(fs, "/small.txt", []byte("match"), 0644))

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "match",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		assert.Len(t, matches, 1)
		assert.Equal(t, "/small.txt", matches[0]["file"])
	})

	t.Run("Match Limit", func(t *testing.T) {
		fs, toolDef := setup()
		// Create a file with 150 matches
		var content strings.Builder
		for i := 0; i < 150; i++ {
			content.WriteString("match\n")
		}
		require.NoError(t, afero.WriteFile(fs, "/many.txt", []byte(content.String()), 0644))

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "match",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		assert.Len(t, matches, 100)
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		fs, toolDef := setup()
		// Create many files to ensure searching takes some "time" (iterations)
		for i := 0; i < 100; i++ {
			require.NoError(t, afero.WriteFile(fs, fmt.Sprintf("/file%d.txt", i), []byte("match"), 0644))
		}

		ctx, cancel := context.WithCancel(context.Background())
		// Cancel immediately
		cancel()

		args := map[string]interface{}{
			"path":    "/",
			"pattern": "match",
		}
		_, err := toolDef.Handler(ctx, args)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("Relative Path Calculation", func(t *testing.T) {
		// Test that the returned file path is correctly joined with the input path
		fs, toolDef := setup()
		require.NoError(t, fs.MkdirAll("/data/subdir", 0755))
		require.NoError(t, afero.WriteFile(fs, "/data/subdir/file.txt", []byte("match"), 0644))

		// Input path: /data
		args := map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		}
		res, err := toolDef.Handler(context.Background(), args)
		require.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})

		assert.Len(t, matches, 1)
		// Expected: /data/subdir/file.txt
		assert.Equal(t, filepath.Join("/data", "subdir/file.txt"), matches[0]["file"])
	})
}
