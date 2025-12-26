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

// setupUpstreamHelper creates a filesystem upstream and returns the tool manager and service ID.
func setupUpstreamHelper(t *testing.T, tempDir string, ro bool) (tool.ManagerInterface, string) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_additional"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				ReadOnly: proto.Bool(ro),
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
	return tm, id
}

func findToolHelper(tm tool.ManagerInterface, id, name string) tool.Tool {
	tool, ok := tm.GetTool(id + "." + name)
	if ok {
		return tool
	}
	return nil
}

func TestFilesystemUpstream_SearchFiles_Advanced(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_search_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tm, id := setupUpstreamHelper(t, tempDir, false)
	searchTool := findToolHelper(tm, id, "search_files")
	require.NotNil(t, searchTool)

	// 1. Test skipping hidden directories
	err = os.Mkdir(filepath.Join(tempDir, ".hidden"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, ".hidden", "secret.txt"), []byte("find me hidden"), 0644)
	require.NoError(t, err)

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
	// Should not find the one in .hidden
	for _, m := range matches {
		assert.NotContains(t, m["file"], ".hidden")
	}

	// 2. Test skipping binary files
	// Create a "binary" file (null bytes)
	binaryContent := []byte("find me binary\x00\x00\x00")
	err = os.WriteFile(filepath.Join(tempDir, "binary.bin"), binaryContent, 0644)
	require.NoError(t, err)

	res, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "find me",
		},
	})
	require.NoError(t, err)
	resMap = res.(map[string]interface{})
	matches = resMap["matches"].([]map[string]interface{})
	for _, m := range matches {
		assert.NotEqual(t, "binary.bin", m["file"])
	}

	// 3. Test skipping large files
	// Create a large file > 10MB.
	// We can use a sparse file to be fast.
	largeFile := filepath.Join(tempDir, "large.txt")
	f, err := os.Create(largeFile)
	require.NoError(t, err)
	// Seek to 11MB
	_, err = f.Seek(11*1024*1024, 0)
	require.NoError(t, err)
	_, err = f.Write([]byte("find me large"))
	require.NoError(t, err)
	f.Close()

	res, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "find me",
		},
	})
	require.NoError(t, err)
	resMap = res.(map[string]interface{})
	matches = resMap["matches"].([]map[string]interface{})
	for _, m := range matches {
		assert.NotEqual(t, "large.txt", m["file"])
	}

	// 4. Test regex error
	_, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path":    "/data",
			"pattern": "[", // Invalid regex
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestFilesystemUpstream_ValidatePath_MultipleRoots(t *testing.T) {
	tempDir1, err := os.MkdirTemp("", "fs_root1")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir1)
	err = os.WriteFile(filepath.Join(tempDir1, "file1.txt"), []byte("content1"), 0644)
	require.NoError(t, err)

	tempDir2, err := os.MkdirTemp("", "fs_root2")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir2)
	err = os.WriteFile(filepath.Join(tempDir2, "file2.txt"), []byte("content2"), 0644)
	require.NoError(t, err)

	// Configure with multiple roots, one being a prefix of another or just multiple
	// Case: /data -> tempDir1, /data/logs -> tempDir2
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_multiroot"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data":      tempDir1,
					"/data/logs": tempDir2,
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
	readTool := findToolHelper(tm, id, "read_file")

	// Read from /data/file1.txt -> tempDir1/file1.txt
	res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/data/file1.txt",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	assert.Equal(t, "content1", resMap["content"])

	// Read from /data/logs/file2.txt -> tempDir2/file2.txt (Longest prefix match)
	res, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/data/logs/file2.txt",
		},
	})
	require.NoError(t, err)
	resMap = res.(map[string]interface{})
	assert.Equal(t, "content2", resMap["content"])
}

func TestFilesystemUpstream_ValidatePath_Symlinks(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_symlink_valid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create real directory
	realDir := filepath.Join(tempDir, "real")
	err = os.Mkdir(realDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(realDir, "file.txt"), []byte("real content"), 0644)
	require.NoError(t, err)

	// Create symlink inside root pointing to another place inside root
	symlinkDir := filepath.Join(tempDir, "link")
	err = os.Symlink("real", symlinkDir) // Relative symlink
	require.NoError(t, err)

	tm, id := setupUpstreamHelper(t, tempDir, false)
	readTool := findToolHelper(tm, id, "read_file")

	// Read via symlink: /data/link/file.txt -> /data/real/file.txt
	res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/data/link/file.txt",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	assert.Equal(t, "real content", resMap["content"])
}

func TestFilesystemUpstream_CreateFilesystem_Unsupported(t *testing.T) {
	u := NewUpstream()
	// Test HTTP
	config := &configv1.FilesystemUpstreamService{
		FilesystemType: &configv1.FilesystemUpstreamService_Http{
			Http: &configv1.HttpFs{},
		},
	}
	// Since createFilesystem is private, we can trigger it via Register
	svcConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_unsupported"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: config,
		},
	}
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	_, _, _, err := u.Register(context.Background(), svcConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http filesystem is not yet supported")

	// Test Zip
	config.FilesystemType = &configv1.FilesystemUpstreamService_Zip{
		Zip: &configv1.ZipFs{},
	}
	_, _, _, err = u.Register(context.Background(), svcConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zip filesystem is not yet supported")

	// Test GCS
	config.FilesystemType = &configv1.FilesystemUpstreamService_Gcs{
		Gcs: &configv1.GcsFs{},
	}
	_, _, _, err = u.Register(context.Background(), svcConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gcs filesystem is not yet supported")

	// Test SFTP
	config.FilesystemType = &configv1.FilesystemUpstreamService_Sftp{
		Sftp: &configv1.SftpFs{},
	}
	_, _, _, err = u.Register(context.Background(), svcConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sftp filesystem is not yet supported")
}

func TestFilesystemUpstream_ValidatePath_RootFallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_root_fallback")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	err = os.WriteFile(filepath.Join(tempDir, "root.txt"), []byte("root"), 0644)
	require.NoError(t, err)

	// Configure with "/" root
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_root_fallback"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/": tempDir,
				},
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
	readTool := findToolHelper(tm, id, "read_file")

	// Access without prefix matching other roots (since only / exists)
	// Virtual path: /some/path/to/file. But wait, if we map / -> tempDir, then /some/path -> tempDir/some/path.
	// We want to test accessing root.txt at /root.txt
	res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/root.txt",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	assert.Equal(t, "root", resMap["content"])
}

func TestFilesystemUpstream_WriteFile_CreateParent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_write_parent")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tm, id := setupUpstreamHelper(t, tempDir, false)
	writeTool := findToolHelper(tm, id, "write_file")

	// Write to deep non-existent subdirectory
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/data/a/b/c/file.txt",
			"content": "deep",
		},
	})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "a/b/c/file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "deep", string(content))
}

func TestFilesystemUpstream_Register_SanitizationError(t *testing.T) {
	u := NewUpstream()
	// Test Sanitization Error
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String(""), // Empty name should fail sanitization
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{},
		},
	}
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
}

func TestFilesystemUpstream_Register_NilService(t *testing.T) {
	u := NewUpstream()
	config := &configv1.UpstreamServiceConfig{
		Name:          proto.String("nil_svc"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{}, // FilesystemService inside is nil if not set? No, wrapper.
	}
	// Actually we need to set the internal field to nil if possible, or trigger the check.
	// The proto generated code usually returns nil if the field is not set.
	// But here ServiceConfig is a oneof.
	// If we provide a Config that is NOT FilesystemService, Register might fail before checking internal nil?
	// Register takes *configv1.UpstreamServiceConfig.
	// Inside: fsService := serviceConfig.GetFilesystemService()
	// If serviceConfig.ServiceConfig is NOT FilesystemService, GetFilesystemService() returns nil.

	config = &configv1.UpstreamServiceConfig{
		Name: proto.String("nil_svc"),
		// No ServiceConfig set
	}
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filesystem service config is nil")
}

func TestFilesystemUpstream_ResolvePath_NonExistent_Symlinks(t *testing.T) {
	// This covers the tricky part of validatePath where target doesn't exist yet
	// and we need to check ancestors for symlinks.
	tempDir, err := os.MkdirTemp("", "fs_resolve_complex")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create structure: /data/real_dir
	err = os.Mkdir(filepath.Join(tempDir, "real_dir"), 0755)
	require.NoError(t, err)

	// Create symlink: /data/link_dir -> /data/real_dir
	err = os.Symlink("real_dir", filepath.Join(tempDir, "link_dir"))
	require.NoError(t, err)

	tm, id := setupUpstreamHelper(t, tempDir, false)
	writeTool := findToolHelper(tm, id, "write_file")

	// Write to /data/link_dir/new_file.txt
	// resolvePath should resolve link_dir -> real_dir and allow it.
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/data/link_dir/new_file.txt",
			"content": "via symlink",
		},
	})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "real_dir", "new_file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "via symlink", string(content))
}

func TestFilesystemUpstream_ResolvePath_Symlink_Escape(t *testing.T) {
	// Test if a symlink in the path points outside root, even if target doesn't exist yet (ancestor check).
	tempDir, err := os.MkdirTemp("", "fs_resolve_escape")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outsideDir, err := os.MkdirTemp("", "outside")
	require.NoError(t, err)
	defer os.RemoveAll(outsideDir)

	// /data/bad_link -> /outside
	err = os.Symlink(outsideDir, filepath.Join(tempDir, "bad_link"))
	require.NoError(t, err)

	tm, id := setupUpstreamHelper(t, tempDir, false)
	writeTool := findToolHelper(tm, id, "write_file")

	// Try to write to /data/bad_link/new_file.txt
	// Ancestor /data/bad_link exists and points to /outside.
	// /outside/new_file.txt is outside root /data.
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/data/bad_link/new_file.txt",
			"content": "should fail",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestFilesystemUpstream_ArgsValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_args")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	tm, id := setupUpstreamHelper(t, tempDir, false)

	// List directory missing path
	lsTool := findToolHelper(tm, id, "list_directory")
	_, err = lsTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "list_directory",
		Arguments: map[string]interface{}{},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")

	// Read file missing path
	readTool := findToolHelper(tm, id, "read_file")
	_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{},
	})
	assert.Error(t, err)

	// Read file path is directory
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	require.NoError(t, err)
	_, err = readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/data/subdir",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is a directory")

	// Write file missing content
	writeTool := findToolHelper(tm, id, "write_file")
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path": "/data/f.txt",
		},
	})
	assert.Error(t, err)

	// Delete file missing path
	delTool := findToolHelper(tm, id, "delete_file")
	_, err = delTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "delete_file",
		Arguments: map[string]interface{}{},
	})
	assert.Error(t, err)

	// Search missing pattern
	searchTool := findToolHelper(tm, id, "search_files")
	_, err = searchTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "search_files",
		Arguments: map[string]interface{}{
			"path": "/data",
		},
	})
	assert.Error(t, err)

	// Get info missing path
	infoTool := findToolHelper(tm, id, "get_file_info")
	_, err = infoTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "get_file_info",
		Arguments: map[string]interface{}{},
	})
	assert.Error(t, err)
}

func TestFilesystemUpstream_ValidatePath_NoRoots(t *testing.T) {
	// Manually construct config with no roots to test error in validatePath
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_noroot"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{}, // Empty
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

	// Any operation should fail
	lsTool := findToolHelper(tm, id, "list_directory")
	_, err = lsTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "list_directory",
		Arguments: map[string]interface{}{
			"path": "/data",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no root paths defined")
}

func TestFilesystemUpstream_ValidatePath_NoMatchingRoot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_nomatch")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_nomatch"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
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

	lsTool := findToolHelper(tm, id, "list_directory")
	// /other is not under /data
	_, err = lsTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "list_directory",
		Arguments: map[string]interface{}{
			"path": "/other",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestFilesystemUpstream_ValidatePath_RootResolutionFail(t *testing.T) {
	// Test case where root path does not exist on disk
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_badroot"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": "/non/existent/path/on/system",
				},
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

	lsTool := findToolHelper(tm, id, "list_directory")
	_, err = lsTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "list_directory",
		Arguments: map[string]interface{}{
			"path": "/data",
		},
	})
	assert.Error(t, err)
	// Error message depends on OS but should fail resolving root
}

func TestFilesystemUpstream_ResolvePath_DefaultType(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_default_type")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	err = os.WriteFile(filepath.Join(tempDir, "default.txt"), []byte("default"), 0644)
	require.NoError(t, err)

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_default_type"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPaths: map[string]string{
					"/data": tempDir,
				},
				// FilesystemType is nil
			},
		},
	}

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	readTool := findToolHelper(tm, id, "read_file")
	res, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"path": "/data/default.txt",
		},
	})
	require.NoError(t, err)
	resMap := res.(map[string]interface{})
	assert.Equal(t, "default", resMap["content"])
}
