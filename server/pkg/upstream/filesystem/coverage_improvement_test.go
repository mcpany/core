package filesystem

import (
	"context"
	"fmt"
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

func TestCreateProvider_Coverage(t *testing.T) {
	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	// Test HTTP filesystem (unsupported)
	configHttp := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_http"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			Http: configv1.HttpFs_builder{Endpoint: proto.String("http://example.com")}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(context.Background(), configHttp, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http filesystem is not yet supported")

	// Test GCS filesystem (creation failure due to missing creds/config)
	configGcs := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_gcs"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			Gcs: configv1.GcsFs_builder{Bucket: proto.String("mybucket")}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(context.Background(), configGcs, tm, nil, nil, false)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create filesystem provider")
	}

	// Test SFTP filesystem (fail)
	configSftp := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_sftp"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			Sftp: configv1.SftpFs_builder{
				Address:  proto.String("invalid.host.local:2222"),
				Username: proto.String("user"),
				Password: proto.String("pass"),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(context.Background(), configSftp, tm, nil, nil, false)
	assert.Error(t, err)

	// Test Zip filesystem (fail due to missing file)
	configZip := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_zip"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			Zip: configv1.ZipFs_builder{FilePath: proto.String("non_existent.zip")}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(context.Background(), configZip, tm, nil, nil, false)
	assert.Error(t, err)

	// Test S3 filesystem (might succeed creation but cover the case)
	configS3 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_s3"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			S3: configv1.S3Fs_builder{
				Bucket: proto.String("mybucket"),
				Region: proto.String("us-east-1"),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = u.Register(context.Background(), configS3, tm, nil, nil, false)
	// NewS3Provider might return error if config is invalid (e.g. nil), but here it's valid enough.
}

func TestCall_Coverage(t *testing.T) {
	// Test Call method of fsCallable directly to cover unmarshal error
	handler := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}
	c := &fsCallable{handler: handler}

	// Invalid JSON
	req := &tool.ExecutionRequest{
		ToolInputs: []byte("{invalid json"),
	}
	_, err := c.Call(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal arguments")
}

func TestTools_ErrorPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tools_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_error"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/data": tempDir,
			},
			DeniedPaths: []string{
				filepath.Join(tempDir, "denied"),
			},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)
	id, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	findTool := func(name string) tool.Tool {
		tTool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tTool
		}
		return nil
	}

	// Setup denied directory
	deniedDir := filepath.Join(tempDir, "denied")
	err = os.Mkdir(deniedDir, 0755)
	require.NoError(t, err)

	t.Run("move_file_errors", func(t *testing.T) {
		mvTool := findTool("move_file")

		// Source denied
		_, err := mvTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/denied/file",
				"destination": "/data/dest",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")

		// Destination denied
		// Create source file
		err = os.WriteFile(filepath.Join(tempDir, "src"), []byte("data"), 0644)
		require.NoError(t, err)

		_, err = mvTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/src",
				"destination": "/data/denied/dest",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("delete_file_errors", func(t *testing.T) {
		delTool := findTool("delete_file")

		// Path denied
		_, err := delTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "delete_file",
			Arguments: map[string]interface{}{
				"path": "/data/denied/file",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("read_file_errors", func(t *testing.T) {
		rTool := findTool("read_file")

		// Path denied
		_, err := rTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/data/denied/file",
			},
		})
		assert.Error(t, err)

		// Path is directory
		err = os.Mkdir(filepath.Join(tempDir, "somedir"), 0755)
		require.NoError(t, err)
		_, err = rTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{
				"path": "/data/somedir",
			},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is a directory")
	})

	t.Run("missing_args", func(t *testing.T) {
		move := findTool("move_file")
		_, err := move.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "move_file", Arguments: map[string]interface{}{}})
		assert.Error(t, err) // source required

		_, err = move.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "move_file", Arguments: map[string]interface{}{"source": "a"}})
		assert.Error(t, err) // destination required

		del := findTool("delete_file")
		_, err = del.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "delete_file", Arguments: map[string]interface{}{}})
		assert.Error(t, err) // path required

		wr := findTool("write_file")
		_, err = wr.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "write_file", Arguments: map[string]interface{}{}})
		assert.Error(t, err) // path required

		_, err = wr.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "write_file", Arguments: map[string]interface{}{"path": "a"}})
		assert.Error(t, err) // content required
	})

	t.Run("search_files_errors", func(t *testing.T) {
		search := findTool("search_files")
		// Missing path
		_, err := search.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "search_files", Arguments: map[string]interface{}{"pattern": "foo"}})
		assert.Error(t, err)

		// Missing pattern
		_, err = search.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "search_files", Arguments: map[string]interface{}{"path": "/data"}})
		assert.Error(t, err)

		// Invalid regex
		_, err = search.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "search_files", Arguments: map[string]interface{}{"path": "/data", "pattern": "("}})
		assert.Error(t, err)

		// Resolve path error
		_, err = search.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "search_files", Arguments: map[string]interface{}{"path": "/data/denied", "pattern": "foo"}})
		assert.Error(t, err)
	})

	t.Run("write_file_errors", func(t *testing.T) {
		wrTool := findTool("write_file")

		// Create a file
		fPath := filepath.Join(tempDir, "existingfile")
		err := os.WriteFile(fPath, []byte(""), 0644)
		require.NoError(t, err)

		// Try to write to a path where parent is a file (triggers MkdirAll error or ResolvePath error)
		_, err = wrTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/existingfile/target",
				"content": "foo",
			},
		})
		assert.Error(t, err)
		assert.True(t,
			(err != nil && (contains(err.Error(), "failed to create parent directory") || contains(err.Error(), "not a directory"))),
			"error should be either 'failed to create parent directory' or 'not a directory', got: %v", err)

		// Try to write to a directory (triggers WriteFile error)
		dirPath := filepath.Join(tempDir, "existingdir")
		err = os.Mkdir(dirPath, 0755)
		require.NoError(t, err)

		_, err = wrTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "write_file",
			Arguments: map[string]interface{}{
				"path":    "/data/existingdir",
				"content": "foo",
			},
		})
		assert.Error(t, err)
	})

	t.Run("move_file_mkdir_error", func(t *testing.T) {
		mvTool := findTool("move_file")

		// Create a file
		fPath := filepath.Join(tempDir, "conflict")
		err := os.WriteFile(fPath, []byte(""), 0644)
		require.NoError(t, err)

		srcPath := filepath.Join(tempDir, "movetest")
		err = os.WriteFile(srcPath, []byte(""), 0644)
		require.NoError(t, err)

		// Try to move to a path where parent is a file
		_, err = mvTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/movetest",
				"destination": "/data/conflict/target",
			},
		})
		assert.Error(t, err)
		assert.True(t,
			(err != nil && (contains(err.Error(), "failed to create parent directory") || contains(err.Error(), "not a directory"))),
			"error should be either 'failed to create parent directory' or 'not a directory', got: %v", err)
	})

	t.Run("Register_SanitizeError", func(t *testing.T) {
		// Provide empty name
		configInvalid := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(""),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				Tmpfs: configv1.MemMapFs_builder{}.Build(),
			}.Build(),
		}.Build()
		_, _, _, err := u.Register(context.Background(), configInvalid, tm, nil, nil, false)
		assert.Error(t, err)
	})

	t.Run("list_directory_errors", func(t *testing.T) {
		lsTool := findTool("list_directory")
		// Path denied
		_, err := lsTool.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "list_directory", Arguments: map[string]interface{}{"path": "/data/denied"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")

		// Path is file (not directory)
		_, err = lsTool.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "list_directory", Arguments: map[string]interface{}{"path": "/data/existingfile"}})
		assert.Error(t, err)
	})

	t.Run("get_file_info_errors", func(t *testing.T) {
		infoTool := findTool("get_file_info")
		// Path denied
		_, err := infoTool.Execute(context.Background(), &tool.ExecutionRequest{ToolName: "get_file_info", Arguments: map[string]interface{}{"path": "/data/denied"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("Register_DefaultProvider", func(t *testing.T) {
		configNil := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test_nil_fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{
					"/": tempDir,
				},
			}.Build(),
		}.Build()
		_, _, _, err := u.Register(context.Background(), configNil, tm, nil, nil, false)
		require.NoError(t, err)
	})

	t.Run("Register_WrongServiceType", func(t *testing.T) {
		configGrpc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test_grpc"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: proto.String("127.0.0.1:50051"),
			}.Build(),
		}.Build()
		_, _, _, err := u.Register(context.Background(), configGrpc, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "filesystem service config is nil")
	})

	t.Run("search_files_skips", func(t *testing.T) {
		search := findTool("search_files")

		// Create binary file
		binPath := filepath.Join(tempDir, "bin.dat")
		os.WriteFile(binPath, []byte{0, 1, 2, 3}, 0644)

		res, err := search.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{"path": "/data", "pattern": ".*"},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// bin.dat should NOT be in matches
		for _, m := range matches {
			assert.NotEqual(t, "bin.dat", filepath.Base(m["file"].(string)))
		}
	})

	t.Run("move_file_rename_error", func(t *testing.T) {
		mvTool := findTool("move_file")

		dirPath := filepath.Join(tempDir, "movedir")
		os.Mkdir(dirPath, 0755)

		destPath := filepath.Join(tempDir, "destfile")
		os.WriteFile(destPath, []byte("data"), 0644)

		_, err := mvTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "move_file",
			Arguments: map[string]interface{}{
				"source":      "/data/movedir",
				"destination": "/data/destfile",
			},
		})
		assert.Error(t, err)
	})

	t.Run("search_files_max_matches", func(t *testing.T) {
		search := findTool("search_files")

		// Create 105 files
		subdir := filepath.Join(tempDir, "manyfiles")
		os.Mkdir(subdir, 0755)
		for i := 0; i < 105; i++ {
			os.WriteFile(filepath.Join(subdir, fmt.Sprintf("file_%d.txt", i)), []byte("match me"), 0644)
		}

		res, err := search.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{"path": "/data/manyfiles", "pattern": "match me"},
		})
		require.NoError(t, err)
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})

		// Should be limited to 100
		assert.Len(t, matches, 100)
	})

	t.Run("search_files_cancellation", func(t *testing.T) {
		search := findTool("search_files")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := search.Execute(ctx, &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{"path": "/data", "pattern": "foo"},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Register_Twice_HealthChecker", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test_re_register"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				Tmpfs: configv1.MemMapFs_builder{}.Build(),
			}.Build(),
		}.Build()

		_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
		require.NoError(t, err)

		// Register again to trigger checker stop
		_, _, _, err = u.Register(context.Background(), config, tm, nil, nil, false)
		require.NoError(t, err)
	})

	t.Run("read_file_permission_error", func(t *testing.T) {
		readTool := findTool("read_file")

		fPath := filepath.Join(tempDir, "unreadable")
		os.WriteFile(fPath, []byte("secret"), 0000)
		// Ensure cleanup can remove it
		defer os.Chmod(fPath, 0644)

		_, err := readTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "read_file",
			Arguments: map[string]interface{}{"path": "/data/unreadable"},
		})
		assert.Error(t, err)
	})

	t.Run("search_files_unreadable", func(t *testing.T) {
		// Search should skip unreadable files without error
		search := findTool("search_files")

		fPath := filepath.Join(tempDir, "unreadable_search")
		os.WriteFile(fPath, []byte("secret"), 0000)
		defer os.Chmod(fPath, 0644)

		res, err := search.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName: "search_files",
			Arguments: map[string]interface{}{"path": "/data", "pattern": "secret"},
		})
		require.NoError(t, err)
		// Should not match unreadable file
		resMap := res.(map[string]interface{})
		matches := resMap["matches"].([]map[string]interface{})
		for _, m := range matches {
			assert.NotEqual(t, "unreadable_search", filepath.Base(m["file"].(string)))
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && len(substr) > 0 &&
		(s[0:len(substr)] == substr || contains(s[1:], substr))
}

func TestValidateLocalPaths_Error(t *testing.T) {
	// Create a directory with no permissions to trigger os.Stat error
	tempDir, err := os.MkdirTemp("", "validate_error")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	lockedDir := filepath.Join(tempDir, "locked")
	err = os.Mkdir(lockedDir, 0000) // No permissions
	require.NoError(t, err)

	// Use a path inside locked dir as root
	targetPath := filepath.Join(lockedDir, "subdir")

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test_validate_error"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{
				"/locked": targetPath,
			},
			Os: configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	// Register should warn but not fail?
	// validateLocalPaths just warns and deletes from map.
	_, _, _, err = u.Register(context.Background(), config, tm, nil, nil, false)
	require.NoError(t, err)

	// Clean up permissions so os.RemoveAll works
	os.Chmod(lockedDir, 0755)
}
