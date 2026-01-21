// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

// TestHTTPQueryOverrideBug reproduces a bug where a query parameter in the tool definition
// fails to override a base query parameter if the tool parameter value has invalid percent encoding.
//
// Expected behavior: The tool parameter should override the base parameter.
// Actual behavior (before fix): Both parameters are present in the final URL.
func TestHTTPQueryOverrideBug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	toolManager := tool.NewMockManagerInterface(ctrl)
	promptManager := prompt.NewMockManagerInterface(ctrl)
	resourceManager := resource.NewMockManagerInterface(ctrl)
	poolManager := pool.NewManager()

	u := http.NewUpstream(poolManager)

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://example.com/api?q=default"),
				Tools: []*configv1.ToolDefinition{
					{
						Name:   proto.String("test-tool"),
						CallId: proto.String("test-call"),
					},
				},
				Calls: map[string]*configv1.HttpCallDefinition{
					"test-call": {
						Id:           proto.String("test-call"),
						EndpointPath: proto.String("/foo?q=%GG"), // Invalid percent encoding
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					},
				},
			},
		},
	}

	// Expect AddServiceInfo
	toolManager.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any()).AnyTimes()
	toolManager.EXPECT().ClearToolsForService(gomock.Any()).AnyTimes()

	// Capture the tool to check its definition
	var capturedTool tool.Tool
	toolManager.EXPECT().AddTool(gomock.Any()).DoAndReturn(func(t tool.Tool) error {
		capturedTool = t
		return nil
	}).Times(1)

	_, _, _, err := u.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, true)
	require.NoError(t, err)
	require.NotNil(t, capturedTool)

	fqn := capturedTool.Tool().GetUnderlyingMethodFqn()
	// Check the URL in the FQN (Format: "GET http://...")
	parts := strings.SplitN(fqn, " ", 2)
	require.Len(t, parts, 2)
	url := parts[1]

	// The bug is that we get "q=default&q=%GG"
	// We want just "q=%GG" (because tool param overrides base param)

	// Check if base param is present
	assert.NotContains(t, url, "q=default", "Base query parameter should be overridden")
	assert.Contains(t, url, "q=%GG", "Tool query parameter should be present")
}
