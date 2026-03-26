// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestManager_AddTool_EnforcesNamespacing(t *testing.T) {
	tm := NewManager(nil)

	// Case 1: Tool with ServiceID
	t1Proto := v1.Tool_builder{
		Name:      proto.String("shortname"),
		ServiceId: proto.String("myservice"),
	}.Build()
	t1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return t1Proto
		},
	}

	err := tm.AddTool(t1)
	require.NoError(t, err)

	// Verify short name is NOT registered
	_, found := tm.GetTool("shortname")
	assert.False(t, found, "Short name 'shortname' should NOT be findable when ServiceID is present")

	// Verify fully qualified name IS registered
	_, found = tm.GetTool("myservice.shortname")
	assert.True(t, found, "Fully qualified name 'myservice.shortname' SHOULD be findable")
}
