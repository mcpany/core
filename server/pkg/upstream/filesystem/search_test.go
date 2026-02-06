// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFiles(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		// Helper
		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		createFile("/data/file1.txt", "hello world\nthis is a test")
		createFile("/data/file2.txt", "another file")
		createFile("/data/subdir/file3.txt", "test match inside")

		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 2)

		found := make(map[string]bool)
		for _, m := range matches {
			found[m["file"].(string)] = true
			if m["file"] == "/data/file1.txt" {
				assert.Equal(t, 2, m["line_number"])
				assert.Equal(t, "this is a test", m["line_content"])
			}
		}
		assert.True(t, found["/data/file1.txt"])
		assert.True(t, found["/data/subdir/file3.txt"])
	})

	t.Run("InputValidation", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		// Missing path
		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"pattern": "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")

		// Missing pattern
		_, err = toolDef.Handler(context.Background(), map[string]interface{}{
			"path": "/data",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pattern is required")
	})

	t.Run("RegexValidation", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		_, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "[", // Invalid regex
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("Exclusions", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		createFile("/data/src/main.js", "console.log('test')")
		createFile("/data/src/my.test.js", "console.log('test')")
		createFile("/data/node_modules/pkg/index.js", "console.log('test')")
		createFile("/data/dist/bundle.js", "console.log('test')")

		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
			"exclude_patterns": []interface{}{
				"*.test.js",
				"node_modules",
				"dist",
			},
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/data/src/main.js", matches[0]["file"])
	})

	t.Run("HiddenDirectories", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		// Files in hidden directory should be skipped
		createFile("/data/.git/config", "test config")
		createFile("/data/visible.txt", "visible test")

		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/data/visible.txt", matches[0]["file"])
	})

	t.Run("BinaryFiles", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		// Create a file with strictly binary content at start
		// http.DetectContentType needs non-text control chars to detect octet-stream generally
		binaryContent := []byte{0x00, 0x01, 0x02, 0x03}
		binaryContent = append(binaryContent, []byte("test match")...)
		_ = afero.WriteFile(fs, "/data/binary.bin", binaryContent, 0644)
		createFile("/data/text.txt", "test match")

		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/data/text.txt", matches[0]["file"])
	})

	t.Run("LargeFiles", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		largeContent := make([]byte, 10*1024*1024+1)
		copy(largeContent, []byte("test match"))

		_ = afero.WriteFile(fs, "/data/large.txt", largeContent, 0644)
		createFile("/data/small.txt", "test match")

		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, "/data/small.txt", matches[0]["file"])
	})

	t.Run("MaxMatches", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		// Create a file with 105 matches
		content := ""
		for i := 0; i < 105; i++ {
			content += "test match\n"
		}
		createFile("/data/many.txt", content)

		res, err := toolDef.Handler(context.Background(), map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
		})
		require.NoError(t, err)

		matches := res["matches"].([]map[string]interface{})
		assert.Equal(t, 100, len(matches))
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		p := provider.NewTmpfsProvider()
		defer p.Close()
		fs := p.GetFs()
		toolDef := searchFilesTool(p, fs)

		createFile := func(path, content string) {
			dir := filepath.Dir(path)
			_ = fs.MkdirAll(dir, 0755)
			_ = afero.WriteFile(fs, path, []byte(content), 0644)
		}

		createFile("/data/slow.txt", "test match")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := toolDef.Handler(ctx, map[string]interface{}{
			"path":    "/data",
			"pattern": "test",
		})
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}
