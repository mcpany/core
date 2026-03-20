// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPIUpstream_Register_SpaceInPath_Bug(t *testing.T) {
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	specContent := `
openapi: 3.0.0
info:
  title: Space API
  version: 1.0.0
paths:
  /path with space:
    get:
      operationId: getPathWithSpace
      responses:
        '200':
          description: OK
`

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("space-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecContent: proto.String(specContent),
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("space-service")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
	mockToolManager.On("GetTool", mock.Anything).Return(nil, false)

    // We expect AddTool to be called. If the bug exists, it won't be called.
    var addedTool tool.Tool
	mockToolManager.On("AddTool", mock.Anything).Run(func(args mock.Arguments) {
		addedTool = args.Get(0).(tool.Tool)
	}).Return(nil)

	_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)

    // The bug: strings.Fields splits "/path with space" into parts, causing "Invalid underlying method FQN" error
    // so the tool is skipped and not registered.

    if addedTool == nil {
        t.Fatal("Bug reproduced: Tool was not registered due to space in path")
    }

    // Verify the FQN
	fqn := addedTool.Tool().GetUnderlyingMethodFqn()
	assert.Contains(t, fqn, "/path with space")
}
