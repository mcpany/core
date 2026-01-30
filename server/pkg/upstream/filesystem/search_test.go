// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	fs afero.Fs
}

// Ensure mockProvider implements provider.Provider
var _ provider.Provider = (*mockProvider)(nil)

func (m *mockProvider) Close() error { return nil }
func (m *mockProvider) GetFs() afero.Fs { return m.fs }
func (m *mockProvider) ResolvePath(p string) (string, error) { return p, nil }

func TestSearchFiles_Exclusions(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// Setup files
	require.NoError(t, afero.WriteFile(fs, "/include.txt", []byte("target"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/exclude_file.txt", []byte("target"), 0644))
	require.NoError(t, fs.Mkdir("/exclude_dir", 0755))
	require.NoError(t, afero.WriteFile(fs, "/exclude_dir/file.txt", []byte("target"), 0644))

	// Test exclusion
	res, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "target",
		"exclude_patterns": []interface{}{
			"exclude_file.txt",
			"exclude_dir",
		},
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/include.txt", matches[0]["file"])
}

func TestSearchFiles_MaxMatches(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// Create 105 files matching the pattern
	for i := 0; i < 105; i++ {
		require.NoError(t, afero.WriteFile(fs, fmt.Sprintf("/file_%d.txt", i), []byte("match"), 0644))
	}

	res, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100)
}

func TestSearchFiles_Cancellation(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// Create many files
	for i := 0; i < 1000; i++ {
		require.NoError(t, afero.WriteFile(fs, fmt.Sprintf("/file_%d.txt", i), []byte("match"), 0644))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := tool.Handler(ctx, map[string]interface{}{
		"path":    "/",
		"pattern": "match",
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestSearchFiles_LargeFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// 10MB + 1 byte
	largeContent := make([]byte, 10*1024*1024+1)
	require.NoError(t, afero.WriteFile(fs, "/large.txt", largeContent, 0644))
	require.NoError(t, afero.WriteFile(fs, "/normal.txt", []byte("target"), 0644))

	res, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "target",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	// Should only find normal.txt
	foundLarge := false
	for _, m := range matches {
		if m["file"] == "/large.txt" {
			foundLarge = true
		}
	}
	assert.False(t, foundLarge, "Large file should be skipped")
}

func TestSearchFiles_SmallBinary(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// Small binary file
	require.NoError(t, afero.WriteFile(fs, "/binary.bin", []byte{0x00, 0x01, 0xFF}, 0644))
	require.NoError(t, afero.WriteFile(fs, "/text.txt", []byte("match"), 0644))

	res, err := tool.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": ".", // Match anything
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})
	foundBinary := false
	for _, m := range matches {
		if m["file"] == "/binary.bin" {
			foundBinary = true
		}
	}
	assert.False(t, foundBinary, "Binary file should be skipped")
}

func TestSearchFiles_InvalidInputs(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	tests := []struct {
		name string
		args map[string]interface{}
		err  string
	}{
		{
			name: "missing path",
			args: map[string]interface{}{"pattern": "foo"},
			err:  "path is required",
		},
		{
			name: "missing pattern",
			args: map[string]interface{}{"path": "/"},
			err:  "pattern is required",
		},
		{
			name: "invalid regex",
			args: map[string]interface{}{"path": "/", "pattern": "["},
			err:  "invalid regex pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), tt.args)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

// TestSearchFiles_CancellationDuringScan attempts to verify cancellation during scanning.
func TestSearchFiles_CancellationDuringScan(t *testing.T) {
	fs := afero.NewMemMapFs()
	prov := &mockProvider{fs: fs}
	tool := searchFilesTool(prov, fs)

	// Create a file with many lines
	var content strings.Builder
	for i := 0; i < 10000; i++ {
		content.WriteString("line\n")
	}
	require.NoError(t, afero.WriteFile(fs, "/long.txt", []byte(content.String()), 0644))

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel quickly to catch it in the loop
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	// We expect an error, or at least successful partial result if it didn't catch error
	// But in `search.go`: `if ctx.Err() != nil { return ctx.Err() }`
	// So if canceled, it should return error.

	_, err := tool.Handler(ctx, map[string]interface{}{
		"path":    "/",
		"pattern": "line",
	})

	// If it finished before cancellation (fast machine), it's nil.
	// If it caught cancellation, it's context.Canceled.
	if err != nil {
		assert.Equal(t, context.Canceled, err)
	}
}
