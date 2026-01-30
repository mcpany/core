// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDuplicateToolDetection(t *testing.T) {
	b, _ := bus.NewProvider(nil)
	m := tool.NewManager(b)

	// Define a mock tool
	t1 := configv1.ToolDefinition_builder{
		Name:      proto.String("my_tool"),
		ServiceId: proto.String("my_service"),
	}.Build()

	mockTool1, err := tool.NewCallableTool(t1, nil, nil, nil, nil)
	require.NoError(t, err)

	// Add first time
	err = m.AddTool(mockTool1)
	require.NoError(t, err)

	// Add second time (same name, same service)
	err = m.AddTool(mockTool1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate tool detected")
}
