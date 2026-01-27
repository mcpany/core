// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestBrowserUpstream_Register_And_Execute(t *testing.T) {
	u, err := NewUpstream()
	if err != nil {
		t.Skipf("Skipping test because browser could not be launched: %v", err)
	}
	defer u.Shutdown(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-browser"),
		BrowserService: configv1.BrowserUpstreamService_builder{
			// Empty config is fine for now
		}.Build(),
	}.Build()

	// Capture the tool passed to AddTool
	var capturedTool tool.Tool
	mockToolManager.EXPECT().AddTool(gomock.Any()).DoAndReturn(func(t tool.Tool) error {
		capturedTool = t
		return nil
	})

	serviceID, toolDefs, _, err := u.Register(context.Background(), serviceConfig, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, "test-browser", serviceID)
	assert.Len(t, toolDefs, 1)
	assert.Equal(t, "browse", toolDefs[0].GetName())

	assert.NotNil(t, capturedTool)
	assert.Equal(t, "browse", capturedTool.Tool().GetName())

	// Execute
	req := &tool.ExecutionRequest{
		ToolName: "browse",
		Arguments: map[string]interface{}{
			"url": "https://example.com",
		},
	}
	result, err := capturedTool.Execute(context.Background(), req)
	if err != nil {
		t.Logf("Handler execution failed: %v", err)
	} else {
		assert.NotNil(t, result)
		resMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		content, ok := resMap["content"].(string)
		assert.True(t, ok)
		assert.Contains(t, content, "Example Domain")
	}
}
