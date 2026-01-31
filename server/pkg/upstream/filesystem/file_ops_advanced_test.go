// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemUpstream_AdvancedFileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs_adv_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Configure the upstream
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_fs_adv"),
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

	// Register the service
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	// Helper to find a tool by name
	findTool := func(name string) tool.Tool {
		tl, ok := tm.GetTool(id + "." + name)
		if ok {
			return tl
		}
		return nil
	}

	// Test write_file to nested directory (should create parents)
	t.Run("write_file_nested_creation", func(t *testing.T) {
		writeTool := findTool("write_file")
		require.NotNil(t, writeTool)

		// Write to a deeply nested path that doesn't exist
		nestedPath := "/data/a/b/c/nested.txt"
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    nestedPath,
				"content": "nested content",
			},
		})
		require.NoError(t, err)

		// Verify file content
		content, err := os.ReadFile(filepath.Join(tempDir, "a/b/c/nested.txt"))
		require.NoError(t, err)
		assert.Equal(t, "nested content", string(content))
	})

	// Test move_file to nested directory (should create parents)
	t.Run("move_file_nested_creation", func(t *testing.T) {
		moveTool := findTool("move_file")
		require.NotNil(t, moveTool)
		writeTool := findTool("write_file")

		// Create source file
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/move_source.txt",
				"content": "moving content",
			},
		})
		require.NoError(t, err)

		// Move to nested destination
		destPath := "/data/x/y/z/moved.txt"
		_, err = moveTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/move_source.txt",
				"destination": destPath,
			},
		})
		require.NoError(t, err)

		// Verify destination content
		content, err := os.ReadFile(filepath.Join(tempDir, "x/y/z/moved.txt"))
		require.NoError(t, err)
		assert.Equal(t, "moving content", string(content))

		// Verify source is gone
		_, err = os.Stat(filepath.Join(tempDir, "move_source.txt"))
		assert.True(t, os.IsNotExist(err))
	})

	// Test write_file conflict with existing file as directory
	t.Run("write_file_conflict_file_as_dir", func(t *testing.T) {
		writeTool := findTool("write_file")

		// Create a file "blocker"
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/blocker",
				"content": "i am a file",
			},
		})
		require.NoError(t, err)

		// Try to create a file inside "blocker" (blocker/fail.txt)
		// This should fail because "blocker" is a file, not a directory
		_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/blocker/fail.txt",
				"content": "fail",
			},
		})
		assert.Error(t, err)
		// Error message depends on OS, but usually "not a directory" or similar
	})

	// Test move_file conflict with existing file as directory
	t.Run("move_file_conflict_file_as_dir", func(t *testing.T) {
		writeTool := findTool("write_file")
		moveTool := findTool("move_file")

		// Create source file
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/to_move.txt",
				"content": "move me",
			},
		})
		require.NoError(t, err)

		// Create a file "blocker_move"
		_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/blocker_move",
				"content": "i am a file",
			},
		})
		require.NoError(t, err)

		// Try to move to "blocker_move/moved.txt"
		_, err = moveTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/to_move.txt",
				"destination": "/data/blocker_move/moved.txt",
			},
		})
		assert.Error(t, err)
	})

	// Test delete_file recursive on non-existent path (should probably fail or succeed? fs.RemoveAll succeeds if path doesn't exist)
	t.Run("delete_file_recursive_non_existent", func(t *testing.T) {
		deleteTool := findTool("delete_file")

		// RemoveAll in Go returns nil if path does not exist.
		// Let's verify if our wrapper changes that.
		// logic: if recursive { if err := fs.RemoveAll(resolvedPath); ... }
		// so it should succeed.

		_, err := deleteTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "delete_file",
			Arguments: map[string]interface{}{
				"path":      "/data/ghost_dir",
				"recursive": true,
			},
		})
		require.NoError(t, err)
	})

	// Test delete_file NON-recursive on directory (should fail)
	t.Run("delete_file_non_recursive_directory", func(t *testing.T) {
		writeTool := findTool("write_file")
		deleteTool := findTool("delete_file")

		// Create a directory by writing a file inside it
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/dir_to_fail_delete/file.txt",
				"content": "content",
			},
		})
		require.NoError(t, err)

		// Try to delete the directory without recursive
		_, err = deleteTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "delete_file",
			Arguments: map[string]interface{}{
				"path":      "/data/dir_to_fail_delete",
				"recursive": false,
			},
		})
		// os.Remove on a non-empty directory fails
		assert.Error(t, err)
	})
}
