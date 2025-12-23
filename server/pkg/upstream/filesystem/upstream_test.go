// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemUpstream_Register_And_Execute(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Configure the upstream
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(false),
			},
		},
	}

	u := NewUpstream()
	// Create a real tool manager with a nil bus for testing
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	// Register the service
	id, tools, resources, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Len(t, tools, 5) // list, read, write, get_info, list_roots
	assert.Empty(t, resources)

	// Helper to find a tool by name
	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	// Test read_file
	t.Run("read_file", func(t *testing.T) {
		readTool := findTool("read_file")
		require.NotNil(t, readTool)

		res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/data/test.txt",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		assert.Equal(t, "hello world", resMap["content"])
	})

	// Test list_directory
	t.Run("list_directory", func(t *testing.T) {
		lsTool := findTool("list_directory")
		require.NotNil(t, lsTool)

		res, err := lsTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "list_directory",
			Arguments: map[string]interface{}{
				"path": "/data",
			},
		})
		require.NoError(t, err)

		resMap := res.(map[string]interface{})
		entries := resMap["entries"].([]interface{})
		assert.Len(t, entries, 2) // test.txt, subdir
	})

	// Test write_file
	t.Run("write_file", func(t *testing.T) {
		writeTool := findTool("write_file")
		require.NotNil(t, writeTool)

		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/new.txt",
				"content": "new content",
			},
		})
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tempDir, "new.txt"))
		require.NoError(t, err)
		assert.Equal(t, "new content", string(content))
	})

	// Test path traversal (security)
	t.Run("path_traversal", func(t *testing.T) {
		readTool := findTool("read_file")
		require.NotNil(t, readTool)

		_, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/data/../../etc/passwd",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	// Test read-only mode
	t.Run("read_only", func(t *testing.T) {
		// Re-register as read-only
		config.GetFilesystemService().ReadOnly = proto.Bool(true)
		tm.ClearToolsForService(id) // Clear previous tools
		u.Register(context.Background(), config, tm, nil, nil, false)

		writeTool := findTool("write_file")
		require.NotNil(t, writeTool)

		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/readonly.txt",
				"content": "fail",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")
	})

	// Test get_file_info
	t.Run("get_file_info", func(t *testing.T) {
		config.GetFilesystemService().ReadOnly = proto.Bool(false)
		tm.ClearToolsForService(id)
		u.Register(context.Background(), config, tm, nil, nil, false)

		infoTool := findTool("get_file_info")
		require.NotNil(t, infoTool)

		res, err := infoTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "get_file_info",
			Arguments: map[string]interface{}{
				"path": "/data/test.txt",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		assert.Equal(t, "test.txt", resMap["name"])
		assert.Equal(t, false, resMap["is_dir"])
	})

	// Test list_allowed_directories
	t.Run("list_allowed_directories", func(t *testing.T) {
		listRootsTool := findTool("list_allowed_directories")
		require.NotNil(t, listRootsTool)

		res, err := listRootsTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "list_allowed_directories",
			Arguments: map[string]interface{}{},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		roots := resMap["roots"].([]string)
		assert.Contains(t, roots, "/data")
	})

	// Test symlink security
	t.Run("symlink_security", func(t *testing.T) {
		// Create a file outside the root
		outsideDir, err := os.MkdirTemp("", "outside")
		require.NoError(t, err)
		defer os.RemoveAll(outsideDir)

		secretFile := filepath.Join(outsideDir, "secret.txt")
		err = os.WriteFile(secretFile, []byte("secret"), 0644)
		require.NoError(t, err)

		// Create a symlink inside the root pointing to the outside file
		symlinkPath := filepath.Join(tempDir, "link_to_secret")
		err = os.Symlink(secretFile, symlinkPath)
		require.NoError(t, err)

		readTool := findTool("read_file")
		require.NotNil(t, readTool)

		// Attempt to read via symlink
		_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/data/link_to_secret",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})
}
