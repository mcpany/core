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

func TestFilesystemUpstream_UnsupportedTypes(t *testing.T) {
	u := NewUpstream()
	tm := tool.NewManager(nil)

	tests := []struct {
		name   string
		config *configv1.FilesystemUpstreamService
		errMsg string
	}{
		{
			name: "Http",
			config: configv1.FilesystemUpstreamService_builder{
				Http: configv1.HttpFs_builder{}.Build(),
			}.Build(),
			errMsg: "http filesystem is not yet supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test_" + tt.name),
				FilesystemService: tt.config,
			}.Build()
			_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestFilesystemUpstream_InputValidation(t *testing.T) {
	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	// Setup a valid service
	tempDir, err := os.MkdirTemp("", "fs_valid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_validation"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": tempDir},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()

	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	tests := []struct {
		toolName string
		args     map[string]interface{}
		errMsg   string
	}{
		{
			toolName: "list_directory",
			args:     map[string]interface{}{},
			errMsg:   "path is required",
		},
		{
			toolName: "read_file",
			args:     map[string]interface{}{},
			errMsg:   "path is required",
		},
		{
			toolName: "write_file",
			args:     map[string]interface{}{},
			errMsg:   "path is required",
		},
		{
			toolName: "write_file",
			args:     map[string]interface{}{"path": "/test.txt"},
			errMsg:   "content is required",
		},
		{
			toolName: "delete_file",
			args:     map[string]interface{}{},
			errMsg:   "path is required",
		},
		{
			toolName: "search_files",
			args:     map[string]interface{}{},
			errMsg:   "path is required",
		},
		{
			toolName: "search_files",
			args:     map[string]interface{}{"path": "/"},
			errMsg:   "pattern is required",
		},
		{
			toolName: "get_file_info",
			args:     map[string]interface{}{},
			errMsg:   "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			tl := findTool(tt.toolName)
			require.NotNil(t, tl)
			_, err := tl.Execute(context.Background(), &tool.ExecutionRequest{
				ToolName:  tt.toolName,
				Arguments: tt.args,
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestFilesystemUpstream_PathResolution(t *testing.T) {
	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	tempDir, err := os.MkdirTemp("", "fs_path")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Case 1: No root paths
	configNoRoot := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_no_root"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()
	id, _, _, err := u.Register(context.Background(), configNoRoot, tm, nil, nil, false)
	require.NoError(t, err)

	readTool, ok := tm.GetTool(id + ".read_file")
	require.True(t, ok)

	_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:  "read_file",
		Arguments: map[string]interface{}{"path": "/test"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no root paths defined")

	// Case 2: No matching root
	configMismatch := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_mismatch"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/app": tempDir},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()
	id2, _, _, err := u.Register(context.Background(), configMismatch, tm, nil, nil, false)
	require.NoError(t, err)

	readTool2, ok := tm.GetTool(id2 + ".read_file")
	require.True(t, ok)

	_, err = readTool2.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:  "read_file",
		Arguments: map[string]interface{}{"path": "/other/test"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestFilesystemUpstream_SearchEdgeCases(t *testing.T) {
	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	tempDir, err := os.MkdirTemp("", "fs_search")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_search_edge"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": tempDir},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	// Create binary file
	binFile := filepath.Join(tempDir, "binary.bin")
	err = os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644)
	require.NoError(t, err)

	// Create hidden directory
	hiddenDir := filepath.Join(tempDir, ".hidden")
	err = os.Mkdir(hiddenDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(hiddenDir, "secret.txt"), []byte("secret"), 0644)
	require.NoError(t, err)

	searchTool, ok := tm.GetTool(id + ".search_files")
	require.True(t, ok)

	// Invalid regex
	_, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:  "search_files",
		Arguments: map[string]interface{}{"path": "/", "pattern": "["},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:  "search_files",
		Arguments: map[string]interface{}{"path": "/", "pattern": ".*"},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})

	// binary.bin should not be in matches
	for _, m := range matches {
		if m["file"] == "binary.bin" {
			t.Errorf("binary file should be ignored")
		}
		if m["file"] == ".hidden/secret.txt" {
			t.Errorf("hidden file should be ignored")
		}
	}
}

func TestFilesystemUpstream_FileOperations(t *testing.T) {
	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	tempDir, err := os.MkdirTemp("", "fs_ops")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_ops"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": tempDir},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	// 1. Read directory as file
	readTool, ok := tm.GetTool(id + ".read_file")
	require.True(t, ok)
	_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:  "read_file",
		Arguments: map[string]interface{}{"path": "/"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is a directory")

	// 2. Delete non-existent file
	deleteTool, ok := tm.GetTool(id + ".delete_file")
	require.True(t, ok)
	_, err = deleteTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:  "delete_file",
		Arguments: map[string]interface{}{"path": "/nonexistent"},
	})
	assert.Error(t, err) // fs.Remove returns error if not exists
}

func TestFilesystemUpstream_SanitizeError(t *testing.T) {
	// Test error when sanitizing service name fails (empty name)
	u := NewUpstream()
	tm := tool.NewManager(nil)
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(""), // Invalid name
		FilesystemService: configv1.FilesystemUpstreamService_builder{}.Build(),
	}.Build()
	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
}

func TestFilesystemUpstream_NilConfig(t *testing.T) {
	u := NewUpstream()
	tm := tool.NewManager(nil)
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_nil"),
	}.Build()
	// config.FilesystemService is nil by default in builder if not set
	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filesystem service config is nil")
}
