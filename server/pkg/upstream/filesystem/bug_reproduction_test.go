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

func TestFilesystemUpstream_WriteFile_RecursiveCreate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs_recursive_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Configure the upstream
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_recursive"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(false),
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	// Helper to find a tool by name
	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	t.Run("write_file_deep_non_existent_subdir", func(t *testing.T) {
		writeTool := findTool("write_file")
		require.NotNil(t, writeTool)

		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/subdir/nested/deep/newfile.txt",
				"content": "deep content",
			},
		})
		require.NoError(t, err)

		// Verify file exists
		content, err := os.ReadFile(filepath.Join(tempDir, "subdir", "nested", "deep", "newfile.txt"))
		require.NoError(t, err)
		assert.Equal(t, "deep content", string(content))
	})
}
