// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFiles_BinaryFile(t *testing.T) {
	// Setup TmpfsProvider
	prov := provider.NewTmpfsProvider()
	baseFs := prov.GetFs()

	// Create a text file
	require.NoError(t, afero.WriteFile(baseFs, "/text.txt", []byte("hello world"), 0644))

	// Create a binary file (contains null byte)
	require.NoError(t, afero.WriteFile(baseFs, "/binary.bin", []byte("hello \x00 world"), 0644))

	toolDef := searchFilesTool(prov, baseFs)

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "hello",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should only find the text file
	assert.Len(t, matches, 1)
	assert.Equal(t, "/text.txt", matches[0]["file"])
}

func TestSearchFiles_HiddenDir(t *testing.T) {
	prov := provider.NewTmpfsProvider()
	baseFs := prov.GetFs()

	// Create .git/config
	require.NoError(t, baseFs.MkdirAll("/.git", 0755))
	require.NoError(t, afero.WriteFile(baseFs, "/.git/config", []byte("secret=true"), 0644))

	// Create normal file
	require.NoError(t, afero.WriteFile(baseFs, "/normal.txt", []byte("secret=false"), 0644))

	toolDef := searchFilesTool(prov, baseFs)

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "secret",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should only find the normal file
	assert.Len(t, matches, 1)
	assert.Equal(t, "/normal.txt", matches[0]["file"])
}

func TestSearchFiles_MatchLimit(t *testing.T) {
	prov := provider.NewTmpfsProvider()
	baseFs := prov.GetFs()

	// Create 110 files
	for i := 0; i < 110; i++ {
		require.NoError(t, afero.WriteFile(baseFs, fmt.Sprintf("/file%d.txt", i), []byte("match me"), 0644))
	}

	toolDef := searchFilesTool(prov, baseFs)

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should be capped at 100
	assert.Len(t, matches, 100)
}

func TestSearchFiles_InvalidRegex(t *testing.T) {
	prov := provider.NewTmpfsProvider()
	baseFs := prov.GetFs()

	toolDef := searchFilesTool(prov, baseFs)

	_, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "[", // Invalid regex
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFiles_ContextCancellation(t *testing.T) {
	prov := provider.NewTmpfsProvider()
	baseFs := prov.GetFs()

	// Create many files to ensure loop runs for a bit
	for i := 0; i < 100; i++ {
		require.NoError(t, afero.WriteFile(baseFs, fmt.Sprintf("/file%d.txt", i), []byte("match me"), 0644))
	}

	toolDef := searchFilesTool(prov, baseFs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := toolDef.Handler(ctx, map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// MockFileInfo implements os.FileInfo
type MockFileInfo struct {
	name string
	size int64
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *MockFileInfo) ModTime() time.Time { return time.Now() }
func (m *MockFileInfo) IsDir() bool        { return false }
func (m *MockFileInfo) Sys() any           { return nil }

// LargeFileFs wraps afero.Fs to report a large size for a specific file
type LargeFileFs struct {
	afero.Fs
}

func (fs *LargeFileFs) Stat(name string) (os.FileInfo, error) {
	if strings.HasSuffix(name, "large.txt") {
		return &MockFileInfo{name: "large.txt", size: 11 * 1024 * 1024}, nil
	}
	return fs.Fs.Stat(name)
}

// We also need to override LstatIfPossible because afero.Walk might use it
func (fs *LargeFileFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if strings.HasSuffix(name, "large.txt") {
		return &MockFileInfo{name: "large.txt", size: 11 * 1024 * 1024}, true, nil
	}
	if lsf, ok := fs.Fs.(afero.Lstater); ok {
		return lsf.LstatIfPossible(name)
	}
	fi, err := fs.Fs.Stat(name)
	return fi, false, err
}

func TestSearchFiles_LargeFile(t *testing.T) {
	// Setup TmpfsProvider for resolution
	prov := provider.NewTmpfsProvider()
	baseFs := prov.GetFs() // This is a MemMapFs

	// Write a small file that pretends to be large
	require.NoError(t, afero.WriteFile(baseFs, "/large.txt", []byte("match me"), 0644))
	require.NoError(t, afero.WriteFile(baseFs, "/normal.txt", []byte("match me"), 0644))

	mockFs := &LargeFileFs{Fs: baseFs}

	toolDef := searchFilesTool(prov, mockFs)

	res, err := toolDef.Handler(context.Background(), map[string]interface{}{
		"path":    "/",
		"pattern": "match me",
	})
	require.NoError(t, err)

	matches := res["matches"].([]map[string]interface{})

	// Should skip large.txt and only find normal.txt
	assert.Len(t, matches, 1)
	assert.Equal(t, "/normal.txt", matches[0]["file"])
}
