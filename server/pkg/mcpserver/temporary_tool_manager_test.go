// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTemporaryToolManager(t *testing.T) {
	m := NewTemporaryToolManager()

	t.Run("ServiceInfo", func(t *testing.T) {
		info := &tool.ServiceInfo{Name: "test-service"}
		m.AddServiceInfo("svc-1", info)

		got, ok := m.GetServiceInfo("svc-1")
		require.True(t, ok)
		assert.Equal(t, info, got)

		_, ok = m.GetServiceInfo("non-existent")
		assert.False(t, ok)
	})

	t.Run("Tool Operations", func(t *testing.T) {
		mockTool := &tool.MockTool{
			ToolFunc: func() *v1.Tool {
				return v1.Tool_builder{
					Name:      proto.String("test-tool"),
					ServiceId: proto.String("svc-1"),
				}.Build()
			},
		}

		err := m.AddTool(mockTool)
		assert.NoError(t, err)

		// GetTool should now succeed
		gotTool, ok := m.GetTool("svc-1.test-tool")
		assert.True(t, ok, "GetTool should return true")
		assert.Equal(t, mockTool, gotTool)

		// ListTools should return the tool
		tools := m.ListTools()
		assert.Len(t, tools, 1)
		assert.Equal(t, mockTool, tools[0])

		// GetToolCountForService should return 1
		count := m.GetToolCountForService("svc-1")
		assert.Equal(t, 1, count)

		count = m.GetToolCountForService("other")
		assert.Equal(t, 0, count)
	})
}
