package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	mcphealth "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemUpstream_UnavailablePath_Repro(t *testing.T) {
	// Create a temporary directory for valid path
	tempDir, err := os.MkdirTemp("", "fs_repro_valid")
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
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// Verify that the invalid path is preserved in config
	fsService := config.GetFilesystemService()
	roots := fsService.GetRootPaths()

	_, ok := roots["/invalid"]
	assert.True(t, ok, "/invalid path should be preserved")
	assert.Equal(t, invalidPath, roots["/invalid"])

	// Find tools
	findTool := func(name string) tool.Tool {
		tool, ok := tm.GetTool(id + "." + name)
		if ok {
			return tool
		}
		return nil
	}

	writeTool := findTool("write_file")
	require.NotNil(t, writeTool)

	// 1. Valid path should still work
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/valid/test.txt",
			"content": "ok",
		},
	})
	assert.NoError(t, err)

	// 2. Invalid path should return a error
	_, err = writeTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "write_file",
		Arguments: map[string]interface{}{
			"path":    "/invalid/test.txt",
			"content": "fail",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve root path symlinks")

	// 3. Health Check should fail
	checker := mcphealth.NewChecker(config)
	require.NotNil(t, checker)

	result := checker.Check(context.Background())
	assert.Equal(t, health.StatusDown, result.Status, "Health check status should be Down")
}
