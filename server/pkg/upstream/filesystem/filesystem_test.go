// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemUpstream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	// Mock prompt/resource managers are not used, can be nil

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "mcpany-fs-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some files
	err = os.WriteFile(filepath.Join(tmpDir, "hello.txt"), []byte("world"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	require.NoError(t, err)

	u := NewUpstream()

	cfg := &configv1.UpstreamServiceConfig{
		Name: proto.String("my-fs"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				RootPath: proto.String(tmpDir),
				ReadOnly: proto.Bool(false),
			},
		},
	}

	// Expectations
	mockToolManager.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any())

	// Expect 4 tools to be added
	var tools []tool.Tool
	mockToolManager.EXPECT().AddTool(gomock.Any()).Times(4).Do(func(t tool.Tool) {
		tools = append(tools, t)
	}).Return(nil)

	_, toolDefs, _, err := u.Register(context.Background(), cfg, mockToolManager, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, toolDefs, 4)

	// Verify tools work
	ctx := context.Background()

	// Find tools
	var listTool, readTool, writeTool tool.Tool
	for _, t := range tools {
		name := t.Tool().GetName()
		if name == "list_files" {
			listTool = t
		} else if name == "read_file" {
			readTool = t
		} else if name == "write_file" {
			writeTool = t
		}
	}

	require.NotNil(t, listTool, "list_files tool not found")
	require.NotNil(t, readTool, "read_file tool not found")
	require.NotNil(t, writeTool, "write_file tool not found")

	// Test list_files
	listParams := map[string]interface{}{"path": "."}
	listJson, _ := json.Marshal(listParams)
	res, err := listTool.Execute(ctx, &tool.ExecutionRequest{ToolInputs: listJson})
	require.NoError(t, err)
	files := res.([]string)
	assert.Contains(t, files, "hello.txt")
	assert.Contains(t, files, "subdir/")

	// Test read_file
	readParams := map[string]interface{}{"path": "hello.txt"}
	readJson, _ := json.Marshal(readParams)
	res, err = readTool.Execute(ctx, &tool.ExecutionRequest{ToolInputs: readJson})
	require.NoError(t, err)
	assert.Equal(t, "world", res.(string))

	// Test write_file
	writeParams := map[string]interface{}{"path": "new.txt", "content": "foo"}
	writeJson, _ := json.Marshal(writeParams)
	res, err = writeTool.Execute(ctx, &tool.ExecutionRequest{ToolInputs: writeJson})
	require.NoError(t, err)

	// Verify write
	content, _ := os.ReadFile(filepath.Join(tmpDir, "new.txt"))
	assert.Equal(t, "foo", string(content))

	// Test path traversal
	badParams := map[string]interface{}{"path": "../secret.txt"}
	badJson, _ := json.Marshal(badParams)
	_, err = readTool.Execute(ctx, &tool.ExecutionRequest{ToolInputs: badJson})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}
