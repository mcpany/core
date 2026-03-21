// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSearchFiles_BugRepro(t *testing.T) {
	// Configure the upstream with MemMapFs
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_repro"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/": "/",
			},
			Tmpfs: configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	writeTool := findTool("write_file")
	require.NotNil(t, writeTool)

	// Create a file in a subdirectory: /subdir/target.txt
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/subdir/target.txt",
			"content": "find me",
		},
	})
	require.NoError(t, err)

	// Search in /subdir
	searchTool := findTool("search_files")
	require.NotNil(t, searchTool)

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/subdir",
			"pattern": "find me",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	require.Len(t, matches, 1)

	foundFile := matches[0]["file"].(string)

	// Expect the full path starting with the user provided path
	// The input path was "/subdir", file found at target.txt relative to it.
	// So result should be /subdir/target.txt
	assert.Equal(t, "/subdir/target.txt", foundFile, "foundFile should be '/subdir/target.txt'")

	// Try to read the returned file using the path returned by search
	readTool := findTool("read_file")
	_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": foundFile,
		},
	})

	// This should succeed now
	assert.NoError(t, err, "read_file should succeed when using the path returned by search_files")
}
