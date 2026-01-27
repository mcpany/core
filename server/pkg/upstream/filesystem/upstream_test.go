// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
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
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_fs"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/data": tempDir,
			},
			ReadOnly: proto.Bool(false),
			Os:       configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()

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
		configReadOnly := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test_fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(true),
				Os:       configv1.OsFs_builder{}.Build(),
			}.Build(),
		}.Build()

		tm.ClearToolsForService(id) // Clear previous tools
		u.Register(context.Background(), configReadOnly, tm, nil, nil, false)

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
		configWrite := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test_fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(false),
				Os:       configv1.OsFs_builder{}.Build(),
			}.Build(),
		}.Build()

		tm.ClearToolsForService(id)
		u.Register(context.Background(), configWrite, tm, nil, nil, false)

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
		configSearch := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test_fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(false),
				Os:       configv1.OsFs_builder{}.Build(),
			}.Build(),
		}.Build()

		tm.ClearToolsForService(id)
		u.Register(context.Background(), configSearch, tm, nil, nil, false)

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
		matches := resMap["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)

		match := matches[0]
		assert.Equal(t, "/data/searchable.txt", match["file"])
		assert.Equal(t, "find me here", match["line_content"])
		assert.Equal(t, 2, match["line_number"])
	})

	// Test delete_file
	t.Run("delete_file", func(t *testing.T) {
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
		searchTool := findTool("search_files")
		require.NotNil(t, searchTool)

		// Create files
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
		foundFiles := make(map[string]bool)
		for _, m := range matches {
			foundFiles[m["file"].(string)] = true
		}

		assert.True(t, foundFiles["/data/include.txt"], "include.txt should be found")
		assert.True(t, foundFiles[filepath.Join("/data", "src", "foo.js")], "src/foo.js should be found")
		assert.False(t, foundFiles["/data/exclude.log"], "exclude.log should be excluded")
		assert.False(t, foundFiles[filepath.Join("/data", "node_modules", "foo.js")], "node_modules/foo.js should be excluded")
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
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_memfs"),
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
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_zipfs"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/": "/",
			},
			ReadOnly: proto.Bool(true),
			Zip: configv1.ZipFs_builder{
				FilePath: proto.String(zipPath),
			}.Build(),
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

	// Verify we can't write (it's registered as read-only)
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
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_fs_mixed"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/valid":   tempDir,
				"/invalid": invalidPath,
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

	require.NoError(t, err, "Registration should not fail even if one path is missing")
	assert.NotEmpty(t, id)

	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	// 1. Verify valid path works
	writeTool := findTool("write_file")
	require.NotNil(t, writeTool)

	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/valid/test.txt",
			"content": "ok",
		},
	})
	assert.NoError(t, err, "Writing to valid path should succeed")

	// 2. Verify invalid path fails gracefully
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/invalid/test.txt",
			"content": "fail",
		},
	})
	assert.Error(t, err, "Writing to invalid path should fail")

	// 3. Verify list_allowed_directories DOES show the invalid path (but it fails on access)
	listRootsTool := findTool("list_allowed_directories")
	require.NotNil(t, listRootsTool)

	res, err := listRootsTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "list_allowed_directories",
		Arguments: map[string]interface{}{},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	roots := resMap["roots"].([]string)

	assert.Contains(t, roots, "/valid", "Valid path should be present")
	assert.Contains(t, roots, "/invalid", "Invalid path should be preserved")
}
