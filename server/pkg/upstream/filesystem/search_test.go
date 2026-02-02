// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func setupTestUpstream(t *testing.T) (tool.Tool, string) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "search_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	// Configure the upstream
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/data": tempDir,
			},
			ReadOnly: proto.Bool(false),
			Os:       configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	return searchTool, tempDir
}

func TestSearchFiles_HappyPath(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create a file
	filePath := filepath.Join(tempDir, "hello.txt")
	err := os.WriteFile(filePath, []byte("Hello World\nAnother line\nHello Again"), 0644)
	require.NoError(t, err)

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "Hello",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Len(t, matches, 2)
	assert.Equal(t, "/data/hello.txt", matches[0]["file"])
	assert.Equal(t, 1, matches[0]["line_number"])
	assert.Equal(t, "Hello World", matches[0]["line_content"])
	assert.Equal(t, "/data/hello.txt", matches[1]["file"])
	assert.Equal(t, 3, matches[1]["line_number"])
	assert.Equal(t, "Hello Again", matches[1]["line_content"])
}

func TestSearchFiles_InvalidRegex(t *testing.T) {
	searchTool, _ := setupTestUpstream(t)

	_, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "[", // Invalid regex
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestSearchFiles_ExcludePatterns(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "match.txt"), []byte("foo"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "ignore.log"), []byte("foo"), 0644))

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":             "/data",
			"pattern":          "foo",
			"exclude_patterns": []interface{}{"*.log"},
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/data/match.txt", matches[0]["file"])
}

func TestSearchFiles_ExcludeDirectories(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create directory structure
	nodeModules := filepath.Join(tempDir, "node_modules")
	require.NoError(t, os.Mkdir(nodeModules, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nodeModules, "lib.js"), []byte("foo"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "app.js"), []byte("foo"), 0644))

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":             "/data",
			"pattern":          "foo",
			"exclude_patterns": []interface{}{"node_modules"},
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/data/app.js", matches[0]["file"])
}

func TestSearchFiles_HiddenDirectories(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create hidden directory
	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte("foo"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "visible.txt"), []byte("foo"), 0644))

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "foo",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Len(t, matches, 1)
	assert.Equal(t, "/data/visible.txt", matches[0]["file"])
}

func TestSearchFiles_FileSizeLimit(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create a large file (10MB + 1 byte)
	largeFile := filepath.Join(tempDir, "large.txt")
	f, err := os.Create(largeFile)
	require.NoError(t, err)
	defer f.Close()

	// Seek to 10MB and write one byte to make it sparse but logically large
	_, err = f.Seek(10*1024*1024, 0)
	require.NoError(t, err)
	_, err = f.Write([]byte("a"))
	require.NoError(t, err)

	// Write the pattern at the beginning too, so it would be found if size wasn't checked
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	_, err = f.Write([]byte("findme"))
	require.NoError(t, err)

	// Ensure file is synced/closed properly
	f.Sync()
	f.Close()

	// Verify size
	info, err := os.Stat(largeFile)
	require.NoError(t, err)
	assert.True(t, info.Size() > 10*1024*1024)

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "findme",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Empty(t, matches)
}

func TestSearchFiles_BinaryDetection(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create binary file
	binFile := filepath.Join(tempDir, "binary.bin")
	// null bytes usually trigger octet-stream detection
	data := make([]byte, 512) // all zeros
	// add the pattern, but surrounded by nulls
	copy(data[0:], []byte("findme"))
	require.NoError(t, os.WriteFile(binFile, data, 0644))

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "findme",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Empty(t, matches)
}

func TestSearchFiles_MaxMatches(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create file with 150 matches
	var sb strings.Builder
	for i := 0; i < 150; i++ {
		sb.WriteString(fmt.Sprintf("match %d\n", i))
	}
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "many.txt"), []byte(sb.String()), 0644))

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "match",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	assert.Len(t, matches, 100)
}

func TestSearchFiles_ContextCancellation(t *testing.T) {
	searchTool, tempDir := setupTestUpstream(t)

	// Create many files to ensure search takes some time
	for i := 0; i < 1000; i++ {
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, fmt.Sprintf("file_%d.txt", i)), []byte("content"), 0644))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := searchTool.Execute(ctx, &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "content",
		},
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
