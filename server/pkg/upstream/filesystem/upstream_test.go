// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
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
	assert.Len(t, tools, 8) // list, read, write, move, delete, search, get_info, list_roots
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

	// Test recursive delete_file
	t.Run("recursive_delete_file", func(t *testing.T) {
		config.GetFilesystemService().ReadOnly = proto.Bool(false)
		tm.ClearToolsForService(id)
		u.Register(context.Background(), config, tm, nil, nil, false)

		deleteTool := findTool("delete_file")
		require.NotNil(t, deleteTool)

		// Create non-empty directory
		dirPath := filepath.Join(tempDir, "todelete")
		err := os.Mkdir(dirPath, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dirPath, "child.txt"), []byte("hi"), 0644)
		require.NoError(t, err)

		// Try non-recursive delete (should fail)
		_, err = deleteTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "delete_file",
			Arguments: map[string]interface{}{
				"path":      "/data/todelete",
				"recursive": false,
			},
		})
		assert.Error(t, err)

		// Try recursive delete
		res, err := deleteTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "delete_file",
			Arguments: map[string]interface{}{
				"path":      "/data/todelete",
				"recursive": true,
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		assert.Equal(t, true, resMap["success"])

		// Verify deletion
		_, err = os.Stat(dirPath)
		assert.True(t, os.IsNotExist(err))
	})

	// Test move_file
	t.Run("move_file", func(t *testing.T) {
		config.GetFilesystemService().ReadOnly = proto.Bool(false)
		tm.ClearToolsForService(id)
		u.Register(context.Background(), config, tm, nil, nil, false)

		moveTool := findTool("move_file")
		require.NotNil(t, moveTool)

		// Create file to move
		src := filepath.Join(tempDir, "move_src.txt")
		err := os.WriteFile(src, []byte("moving"), 0644)
		require.NoError(t, err)

		// Move
		res, err := moveTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/move_src.txt",
				"destination": "/data/moved_dest.txt",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		assert.Equal(t, true, resMap["success"])

		// Verify
		_, err = os.Stat(src)
		assert.True(t, os.IsNotExist(err))
		content, err := os.ReadFile(filepath.Join(tempDir, "moved_dest.txt"))
		require.NoError(t, err)
		assert.Equal(t, "moving", string(content))
	})

	// Test search_files exclusions
	t.Run("search_files_exclusions", func(t *testing.T) {
		config.GetFilesystemService().ReadOnly = proto.Bool(false)
		tm.ClearToolsForService(id)
		u.Register(context.Background(), config, tm, nil, nil, false)

		searchTool := findTool("search_files")
		require.NotNil(t, searchTool)

		// Create files
		// /data/include.txt
		// /data/exclude.log
		// /data/node_modules/foo.js
		// /data/src/foo.js

		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "include.txt"), []byte("match me"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "exclude.log"), []byte("match me"), 0644))

		nodeModules := filepath.Join(tempDir, "node_modules")
		require.NoError(t, os.Mkdir(nodeModules, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(nodeModules, "foo.js"), []byte("match me"), 0644))

		srcDir := filepath.Join(tempDir, "src")
		require.NoError(t, os.Mkdir(srcDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "foo.js"), []byte("match me"), 0644))

		// Search with exclusions
		res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/data",
				"pattern": "match me",
				"exclude_patterns": []interface{}{
					"*.log",
					"node_modules",
				},
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// Should match include.txt and src/foo.js
		// Should NOT match exclude.log and node_modules/foo.js
		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}

		assert.True(t, foundFiles["include.txt"], "include.txt should be found")
		// on windows path separator might differ, but tests run on linux usually
		// filepath.Rel returns OS specific separators.
		// "src/foo.js" might be "src\foo.js" on windows.
		// The test environment is linux based on standard tools.
		assert.True(t, foundFiles[filepath.Join("src", "foo.js")], "src/foo.js should be found")
		assert.False(t, foundFiles["exclude.log"], "exclude.log should be excluded")
		assert.False(t, foundFiles[filepath.Join("node_modules", "foo.js")], "node_modules/foo.js should be excluded")
	})

	// Test read_file binary check
	t.Run("read_file_binary", func(t *testing.T) {
		// Create a binary file
		binFile := filepath.Join(tempDir, "binary.bin")
		// Write null bytes
		err = os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644)
		require.NoError(t, err)

		// Note: The current read_file implementation uses afero.ReadFile which reads everything.
		// It doesn't explicitly block binary files, but search_files does.
		// Let's test search_files with binary.
		searchTool := findTool("search_files")
		res, err := searchTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{
				"path":    "/data",
				"pattern": ".*",
			},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// Should not match binary file
		for _, m := range matches {
			assert.NotEqual(t, "binary.bin", m["file"])
		}
	})

	// Test read_file size limit
	t.Run("read_file_size_limit", func(t *testing.T) {
		// Create a large file (> 10MB)
		largeFile := filepath.Join(tempDir, "large.txt")
		f, err := os.Create(largeFile)
		require.NoError(t, err)
		// Write 10MB + 1 byte
		if err := f.Truncate(10*1024*1024 + 1); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()

		readTool := findTool("read_file")
		_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/data/large.txt",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds limit")
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

func TestFilesystemUpstream_ZipFs(t *testing.T) {
	// Create a temporary zip file
	tempDir, err := os.MkdirTemp("", "zip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Allow tempDir for validation
	validation.SetAllowedPaths([]string{tempDir})
	defer validation.SetAllowedPaths(nil)

	zipPath := filepath.Join(tempDir, "test.zip")
	f, err := os.Create(zipPath)
	require.NoError(t, err)

	w := zip.NewWriter(f)

	// Add file to zip
	f1, err := w.Create("hello.txt")
	require.NoError(t, err)
	_, err = f1.Write([]byte("zip content"))
	require.NoError(t, err)

	// Add subdir to zip
	_, err = w.Create("subdir/")
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)
	f.Close()

	// Configure the upstream with ZipFs
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_zipfs"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/": "/",
				},
				ReadOnly: proto.Bool(true), // Zip is typically read-only or we treat it as such for now
				FilesystemType: &configv1.FilesystemUpstreamService_Zip{
					Zip: &configv1.ZipFs{
						FilePath: proto.String(zipPath),
					},
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

	// Read file from zip
	readTool := findTool("read_file")
	res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/hello.txt",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	assert.Equal(t, "zip content", resMap["content"])

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
	assert.Len(t, entries, 2) // hello.txt, subdir

	// Verify we can't write (it's registered as read-only, and zipfs via afero might be readonly too)
	writeTool := findTool("write_file")
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/new.txt",
			"content": "fail",
		},
	})
	assert.Error(t, err)

	// Clean up
	u.Shutdown(context.Background())
}

func TestFilesystemUpstream_UnavailablePath(t *testing.T) {
	// Create a temporary directory for valid path
	tempDir, err := os.MkdirTemp("", "fs_test_valid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Define an invalid path
	invalidPath := filepath.Join(tempDir, "does_not_exist")

	// Configure the upstream with mixed paths
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_mixed"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/valid":   tempDir,
					"/invalid": invalidPath,
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

	// Register the service
	// We expect this TO fail now
	_, _, _, err = u.Register(context.Background(), config, tm, nil, nil, false)

	require.Error(t, err, "Registration should fail if one path is missing")
	assert.Contains(t, err.Error(), "configured root path does not exist")
}
