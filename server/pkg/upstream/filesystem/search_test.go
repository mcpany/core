// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSearchFiles_HappyPath(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_happy"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	writeTool, ok := tm.GetTool(id + ".write_file")
	require.True(t, ok, "write_file tool not found")
	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	// Setup: Create a file with known content
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/hello.txt",
			"content": "Hello World\nAnother line\nHello again",
		},
	})
	require.NoError(t, err)

	// Execute Search
	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "Hello",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})
	require.Len(t, matches, 2)

	assert.Equal(t, "/hello.txt", matches[0]["file"])
	assert.Equal(t, 1, matches[0]["line_number"])
	assert.Equal(t, "Hello World", matches[0]["line_content"])

	assert.Equal(t, "/hello.txt", matches[1]["file"])
	assert.Equal(t, 3, matches[1]["line_number"])
	assert.Equal(t, "Hello again", matches[1]["line_content"])
}

func TestSearchFiles_InputValidation(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_validation"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	t.Run("Missing Path", func(t *testing.T) {
		_, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"pattern": "foo",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})

	t.Run("Missing Pattern", func(t *testing.T) {
		_, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path": "/",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pattern is required")
	})

	t.Run("Invalid Regex", func(t *testing.T) {
		_, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/",
				"pattern": "[",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})
}

func TestSearchFiles_Exclusions(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_exclusions"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	writeTool, ok := tm.GetTool(id + ".write_file")
	require.True(t, ok, "write_file tool not found")
	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	// Setup files
	files := map[string]string{
		"/match.txt":        "target",
		"/ignore.js":        "target",
		"/nested/match.txt": "target",
		"/skip_dir/foo.txt": "target",
	}

	for path, content := range files {
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    path,
				"content": content,
			},
		})
		require.NoError(t, err)
	}

	// Execute Search with Exclusions
	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "target",
			"exclude_patterns": []interface{}{
				"*.js",
				"skip_dir",
			},
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})

	// Should only find /match.txt and /nested/match.txt
	// Should NOT find /ignore.js or /skip_dir/foo.txt
	assert.Len(t, matches, 2)
	for _, m := range matches {
		file := m["file"].(string)
		if file == "/ignore.js" || strings.Contains(file, "skip_dir") {
			t.Errorf("Found excluded file: %s", file)
		}
	}
}

func TestSearchFiles_HiddenDirectories(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_hidden"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	writeTool, ok := tm.GetTool(id + ".write_file")
	require.True(t, ok, "write_file tool not found")
	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	// Write file in hidden directory
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/.git/config",
			"content": "target",
		},
	})
	require.NoError(t, err)

	// Write file in normal directory
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/visible/config",
			"content": "target",
		},
	})
	require.NoError(t, err)

	// Execute Search
	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "target",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})

	assert.Len(t, matches, 1)
	assert.Equal(t, "/visible/config", matches[0]["file"])
}

func TestSearchFiles_FileConstraints(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_constraints"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	writeTool, ok := tm.GetTool(id + ".write_file")
	require.True(t, ok, "write_file tool not found")
	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	t.Run("Large File", func(t *testing.T) {
		// Create a large string > 10MB
		largeContent := strings.Repeat("a", 10*1024*1024+1)

		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/large.txt",
				"content": largeContent + "target", // target at end
			},
		})
		require.NoError(t, err)

		res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/",
				"pattern": "target",
			},
		})
		require.NoError(t, err)

		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})
		assert.Len(t, matches, 0)
	})

	t.Run("Binary File", func(t *testing.T) {
		// Create file with null bytes
		// write_file takes string, but we can pass string with null bytes
		binaryContent := "\x00\x00\x00\x00target"

		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/binary.bin",
				"content": binaryContent,
			},
		})
		require.NoError(t, err)

		res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/",
				"pattern": "target",
			},
		})
		require.NoError(t, err)

		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})
		assert.Len(t, matches, 0)
	})
}

func TestSearchFiles_Limits(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_limits"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	writeTool, ok := tm.GetTool(id + ".write_file")
	require.True(t, ok, "write_file tool not found")
	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	// Create file with 150 lines containing "target"
	var contentBuilder strings.Builder
	for i := 0; i < 150; i++ {
		contentBuilder.WriteString("target\n")
	}

	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/many_matches.txt",
			"content": contentBuilder.String(),
		},
	})
	require.NoError(t, err)

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "target",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})

	// Max matches is 100
	assert.Len(t, matches, 100)
}

func TestSearchFiles_ContextCancellation(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_cancel"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Tmpfs:     configv1.MemMapFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	writeTool, ok := tm.GetTool(id + ".write_file")
	require.True(t, ok, "write_file tool not found")
	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok, "search_files tool not found")

	// Create a reasonable amount of data
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/file.txt",
			"content": "target",
		},
	})
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = searchTool.Execute(ctx, &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "target",
		},
	})

	// Should return error due to context cancellation
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "context canceled") || err == context.Canceled)
}
