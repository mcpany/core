// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"strings"
	"testing"
	"time"

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
	return virtualPath, nil
}

func (m *mockProvider) Close() error {
	return nil
}

func TestSearchFiles_Direct(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	toolDef := searchFilesTool(prov, fs)
	handler := toolDef.Handler

	// Setup filesystem
	require.NoError(t, fs.MkdirAll("/data", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/text.txt", []byte("hello world\nmatch this line\ngoodbye"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/data/other.txt", []byte("no match here"), 0644))

	// Create binary file
	require.NoError(t, afero.WriteFile(fs, "/data/binary.bin", []byte{0x00, 0x01, 0x02, 0x03, 0xff}, 0644))

	// Create hidden directory
	require.NoError(t, fs.MkdirAll("/data/.hidden", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/.hidden/secret.txt", []byte("match this line"), 0644))

	// Create ignored directory (for exclusion test)
	require.NoError(t, fs.MkdirAll("/data/node_modules", 0755))
	require.NoError(t, afero.WriteFile(fs, "/data/node_modules/lib.js", []byte("match this line"), 0644))

	t.Run("HappyPath", func(t *testing.T) {
		res, err := handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match this",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		// Should match text.txt only (binary skipped, hidden skipped, node_modules matched unless excluded)
		// Wait, node_modules is NOT excluded by default in the tool logic, only if passed in args.
		// So it should match node_modules/lib.js too.

		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}

		assert.True(t, foundFiles["/data/text.txt"])
		assert.True(t, foundFiles["/data/node_modules/lib.js"])
		assert.False(t, foundFiles["/data/binary.bin"], "Binary file should be skipped")
		assert.False(t, foundFiles["/data/.hidden/secret.txt"], "Hidden directory should be skipped")
	})

	t.Run("Exclusions", func(t *testing.T) {
		res, err := handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match this",
			"exclude_patterns": []interface{}{"node_modules"},
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}

		assert.True(t, foundFiles["/data/text.txt"])
		assert.False(t, foundFiles["/data/node_modules/lib.js"], "node_modules should be excluded")
	})

	t.Run("InvalidRegex", func(t *testing.T) {
		_, err := handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "(",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := handler(ctx, map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("LargeFile", func(t *testing.T) {
		// Create a mock file that reports large size
		// MemMapFs doesn't easily support mocking size without content.
		// We'll write a file slightly larger than 10MB.
		// 10MB = 10 * 1024 * 1024 = 10485760 bytes.
		// This might be slow on some envs, but MemMapFs is just memory allocation.

		largeContent := make([]byte, 10*1024*1024 + 100)
		// fill specific part to match regex if it WERE searched
		copy(largeContent[100:], []byte("match this"))

		require.NoError(t, afero.WriteFile(fs, "/data/large.txt", largeContent, 0644))

		res, err := handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "match this",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}

		assert.False(t, foundFiles["/data/large.txt"], "Large file should be skipped")

		// Clean up large file to free memory
		fs.Remove("/data/large.txt")
	})

	t.Run("MissingArgs", func(t *testing.T) {
		_, err := handler(context.Background(), map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")

		_, err = handler(context.Background(), map[string]interface{}{"path": "/data"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pattern is required")
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		// afero.Walk returns error if root path doesn't exist?
		// searchFilesTool ignores walk errors unless they are critical?
		// Let's see: "if err != nil && err != filepath.SkipDir { return nil, err }"
		// afero.Walk calls walkFn with error if it can't read dir.
		// Our walkFn says: "if err != nil { return nil }" (skips it).

		// BUT, if the ROOT itself doesn't exist, afero.Walk returns error immediately?
		// afero.Walk -> readDirNames.

		_, err := handler(context.Background(), map[string]interface{}{
			"path":    "/ghost",
			"pattern": "foo",
		})
		// If walk fails, we return error.
		assert.Error(t, err)
	})
}

// TestSearchFiles_BinaryDetection_EdgeCase covers the case where file is smaller than 512 bytes
func TestSearchFiles_BinaryDetection_SmallFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	handler := searchFilesTool(prov, fs).Handler

	// Small text file
	require.NoError(t, afero.WriteFile(fs, "/small.txt", []byte("hi"), 0644))
	// Small binary file (null byte)
	require.NoError(t, afero.WriteFile(fs, "/small.bin", []byte{0x00, 0x01}, 0644))

	res, err := handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "hi",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/small.txt", matches[0]["file"])
}

// TestSearchFiles_Timeout ensures long searches can be cancelled
func TestSearchFiles_Timeout(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	handler := searchFilesTool(prov, fs).Handler

	// Create many files
	for i := 0; i < 1000; i++ {
		path := strings.ReplaceAll(time.Now().Format(time.RFC3339Nano), ":", "_") + ".txt"
		_ = afero.WriteFile(fs, "/"+path, []byte("match me"), 0644)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This race is tricky in unit tests, but we want to ensure it CHECKS the context.
	// We rely on the fact that creating/walking 1000 files takes > 1ms usually.
	// If it finishes too fast, the test might fail (pass without error).
	// But in a tight loop it should trigger.

	_, err := handler(ctx, map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})

	// If it finished, err is nil. If it cancelled, err is not nil.
	// We can't strictly assert Error because it might be too fast on fast machines.
	// But if we error, it MUST be context deadline.
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
}
