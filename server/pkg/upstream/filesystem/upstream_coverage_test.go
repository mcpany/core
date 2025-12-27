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
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Http{},
			},
			errMsg: "http filesystem is not yet supported",
		},
		{
			name: "Zip",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Zip{},
			},
			errMsg: "zip filesystem is not yet supported",
		},
		{
			name: "Gcs",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Gcs{},
			},
			errMsg: "gcs filesystem is not yet supported",
		},
		{
			name: "Sftp",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Sftp{},
			},
			errMsg: "sftp filesystem is not yet supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configv1.UpstreamServiceConfig{
				Name: proto.String("test_" + tt.name),
				ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
					FilesystemService: tt.config,
				},
			}
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

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_validation"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{"/": tempDir},
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}

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
	configNoRoot := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_no_root"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{},
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}
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
	configMismatch := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_mismatch"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{"/app": tempDir},
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}
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

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_search_edge"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{"/": tempDir},
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}
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

	// Search matching binary content (should be ignored)
	// Note: 0x00 0x01 ...
	// The binary check is: contentType == "application/octet-stream"
	// 4 bytes might be too small to detect?
	// net/http.DetectContentType doc says: "at most the first 512 bytes".
	// Let's check what it detects for 4 bytes starting with null.
	// It should detect octet-stream.

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

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_ops"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{"/": tempDir},
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
			},
		},
	}
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
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String(""), // Invalid name
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{},
		},
	}
	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
}

func TestFilesystemUpstream_NilConfig(t *testing.T) {
	u := NewUpstream()
	tm := tool.NewManager(nil)
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_nil"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: nil,
		},
	}
	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filesystem service config is nil")
}
