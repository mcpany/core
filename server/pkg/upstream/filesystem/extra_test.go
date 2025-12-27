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

func TestFilesystemUpstream_CreateFilesystem_Errors(t *testing.T) {
	u := &Upstream{}

	tests := []struct {
		name        string
		config      *configv1.FilesystemUpstreamService
		expectedErr string
	}{
		{
			name: "Http",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Http{},
			},
			expectedErr: "http filesystem is not yet supported",
		},
		{
			name: "Zip",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Zip{},
			},
			expectedErr: "zip filesystem is not yet supported",
		},
		{
			name: "Gcs",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Gcs{},
			},
			expectedErr: "gcs filesystem is not yet supported",
		},
		{
			name: "Sftp",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Sftp{},
			},
			expectedErr: "sftp filesystem is not yet supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := u.createFilesystem(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestFilesystemUpstream_ValidatePath_EdgeCases(t *testing.T) {
	u := &Upstream{}
	tempDir, err := os.MkdirTemp("", "fs_validate_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	subDir1 := filepath.Join(tempDir, "sub1")
	err = os.Mkdir(subDir1, 0755)
	require.NoError(t, err)

	subDir2 := filepath.Join(tempDir, "sub2")
	err = os.Mkdir(subDir2, 0755)
	require.NoError(t, err)

	rootPaths := map[string]string{
		"/mnt/sub1": subDir1,
		"/mnt":      tempDir,
	}

	// Test longest prefix match
	t.Run("LongestPrefixMatch", func(t *testing.T) {
		// Should match /mnt/sub1
		realPath, err := u.validatePath("/mnt/sub1/file.txt", rootPaths)
		require.NoError(t, err)

		// Resolve symlinks for the directory part since file doesn't exist
		dir, err := filepath.EvalSymlinks(subDir1)
		require.NoError(t, err)
		expected := filepath.Join(dir, "file.txt")

		assert.Equal(t, expected, realPath)
	})

	t.Run("ShorterPrefixMatch", func(t *testing.T) {
		// Should match /mnt
		realPath, err := u.validatePath("/mnt/sub2/file.txt", rootPaths)
		require.NoError(t, err)

		// sub2 exists inside tempDir
		dir, err := filepath.EvalSymlinks(filepath.Join(tempDir, "sub2"))
		require.NoError(t, err)
		expected := filepath.Join(dir, "file.txt")

		assert.Equal(t, expected, realPath)
	})

	t.Run("NoMatch", func(t *testing.T) {
		_, err := u.validatePath("/other/path", rootPaths)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("NoRoots", func(t *testing.T) {
		_, err := u.validatePath("/any", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no root paths defined")
	})
}

func TestFilesystemUpstream_ToolErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_tool_err_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a directory
	dirPath := filepath.Join(tempDir, "mydir")
	err = os.Mkdir(dirPath, 0755)
	require.NoError(t, err)

	// Create a file
	filePath := filepath.Join(tempDir, "myfile.txt")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	require.NoError(t, err)

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_err"),
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

	t.Run("read_file_on_directory", func(t *testing.T) {
		readTool := findTool("read_file")
		_, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/mydir",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is a directory")
	})

	t.Run("read_file_missing_path", func(t *testing.T) {
		readTool := findTool("read_file")
		_, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})

	t.Run("search_files_invalid_regex", func(t *testing.T) {
		searchTool := findTool("search_files")
		_, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/",
				"pattern": "[", // Invalid regex
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("search_files_missing_args", func(t *testing.T) {
		searchTool := findTool("search_files")
		_, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path": "/",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pattern is required")

		_, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"pattern": "foo",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is required")
	})
}

func TestFilesystemUpstream_SearchFiles_BinaryAndHidden(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_search_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 1. Create a binary file (null bytes)
	binFile := filepath.Join(tempDir, "binary.bin")
	// Write enough null bytes and some text that matches
	// We want to ensure it is DETECTED as binary and skipped.
	// content type detection needs first 512 bytes.
	binData := make([]byte, 512)
	// Leave them as 0, which often triggers binary detection or application/octet-stream
	err = os.WriteFile(binFile, binData, 0644)
	require.NoError(t, err)

	// 2. Create a hidden directory with a matching file
	hiddenDir := filepath.Join(tempDir, ".hidden")
	err = os.Mkdir(hiddenDir, 0755)
	require.NoError(t, err)
	hiddenFile := filepath.Join(hiddenDir, "secret.txt")
	err = os.WriteFile(hiddenFile, []byte("find me"), 0644)
	require.NoError(t, err)

	// 3. Create a normal file
	normalFile := filepath.Join(tempDir, "normal.txt")
	err = os.WriteFile(normalFile, []byte("find me"), 0644)
	require.NoError(t, err)

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_search"),
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

	searchTool, _ := tm.GetTool(id + ".search_files")

	res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "find me",
		},
	})
	require.NoError(t, err)

	resMap := res.(map[string]interface{})
	matches := resMap["matches"].([]map[string]interface{})

	// Should match normal.txt
	// Should NOT match binary.bin (even if it had "find me", but here it doesn't.
	// To strictly test binary skip, we should put "find me" in it but make it binary.)
	// Let's rewrite binary file to have "find me" but be binary.

	f, _ := os.OpenFile(binFile, os.O_WRONLY|os.O_TRUNC, 0644)
	f.Write(make([]byte, 100)) // Null bytes to trigger binary
	f.WriteString("find me")
	f.Close()

	// Re-run search
	res, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/",
			"pattern": "find me",
		},
	})
	require.NoError(t, err)
	resMap = res.(map[string]interface{})
	matches = resMap["matches"].([]map[string]interface{})

	// Assertions
	foundNormal := false
	foundHidden := false
	foundBinary := false

	for _, m := range matches {
		file := m["file"].(string)
		if file == "normal.txt" {
			foundNormal = true
		} else if file == ".hidden/secret.txt" { // Should be relative to root
			foundHidden = true
		} else if file == "binary.bin" {
			foundBinary = true
		}
	}

	assert.True(t, foundNormal, "Should find normal file")
	assert.False(t, foundHidden, "Should skip hidden directory")
	assert.False(t, foundBinary, "Should skip binary file")
}
