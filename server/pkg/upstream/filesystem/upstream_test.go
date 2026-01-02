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
				FilesystemType: &configv1.FilesystemUpstreamService_Os{
					Os: &configv1.OsFs{},
				},
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
	assert.Len(t, tools, 7) // list, read, write, delete, search, get_info, list_roots
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

	// Test search_files
	t.Run("search_files", func(t *testing.T) {
		config.GetFilesystemService().ReadOnly = proto.Bool(false)
		tm.ClearToolsForService(id)
		u.Register(context.Background(), config, tm, nil, nil, false)

		searchTool := findTool("search_files")
		require.NotNil(t, searchTool)

		// Create file to search
		writeTool := findTool("write_file")
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/searchable.txt",
				"content": "line one\nfind me here\nline three",
			},
		})
		require.NoError(t, err)

		// Search
		res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/data",
				"pattern": "find me",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		// matches is []map[string]interface{}, not []interface{} because it was constructed that way in Go code.
		// However, when passing through structure conversion or JSON, it might change.
		// In upstream.go: matches := []map[string]interface{}{}
		// So it is of type []map[string]interface{}
		matches := resMap["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)

		match := matches[0]
		assert.Equal(t, "searchable.txt", match["file"])
		assert.Equal(t, "find me here", match["line_content"])
		assert.Equal(t, 2, match["line_number"])
	})

	// Test delete_file
	t.Run("delete_file", func(t *testing.T) {
		config.GetFilesystemService().ReadOnly = proto.Bool(false)
		tm.ClearToolsForService(id)
		u.Register(context.Background(), config, tm, nil, nil, false)

		deleteTool := findTool("delete_file")
		require.NotNil(t, deleteTool)

		// Create file to delete
		writeTool := findTool("write_file")
		_, err := writeTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/deleteme.txt",
				"content": "bye",
			},
		})
		require.NoError(t, err)

		// Delete
		res, err := deleteTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "delete_file",
			Arguments: map[string]interface{}{
				"path": "/data/deleteme.txt",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		assert.Equal(t, true, resMap["success"])

		// Verify deletion
		_, err = os.Stat(filepath.Join(tempDir, "deleteme.txt"))
		assert.True(t, os.IsNotExist(err))
	})

	// Test Shutdown
	t.Run("Shutdown", func(t *testing.T) {
		err := u.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}

func TestFilesystemUpstream_MemMapFs(t *testing.T) {
	// Configure the upstream with MemMapFs
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_memfs"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				// RootPaths are not strictly used for MemMapFs logic, but config validation requires it currently?
				// Let's check upstream.go logic.
				// "if len(fsService.RootPaths) == 0 ... fmt.Errorf("no root paths defined..."
				// We should relax this check if not using OsFs, or provide a dummy one.
				// But wait, MemMapFs is a new filesystem, so it's empty initially.
				// We need to write to it.
				// For MemMapFs, we might want to skip root path validation in Register if we modify Register.
				// But Register still checks `if len(fsService.RootPaths) == 0`.
				// I should fix that in upstream.go first.
				// But for now let's provide a dummy root path to pass validation.
				RootPaths: map[string]string{
					"/": "/",
				},
				FilesystemType: &configv1.FilesystemUpstreamService_Tmpfs{
					Tmpfs: &configv1.MemMapFs{},
				},
			},
		},
	}

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

	// Write a file to MemMapFs
	writeTool := findTool("write_file")
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/hello.txt",
			"content": "memory content",
		},
	})
	require.NoError(t, err)

	// Read it back
	readTool := findTool("read_file")
	res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/hello.txt",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	assert.Equal(t, "memory content", resMap["content"])

	// List directory
	lsTool := findTool("list_directory")
	res, err = lsTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "list_directory",
		Arguments: map[string]interface{}{
			"path": "/",
		},
	})
	require.NoError(t, err)
	resMap = res.(map[string]interface{})
	entries := resMap["entries"].([]interface{})
	assert.Len(t, entries, 1)
	assert.Equal(t, "hello.txt", entries[0].(map[string]interface{})["name"])
}

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
