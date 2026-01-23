// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigErrorTool(t *testing.T) {
	errMsg := "Test error message"
	tool := NewConfigErrorTool(errMsg)

	assert.Equal(t, "mcp_config_error", tool.Tool().GetName())
	assert.Equal(t, "system", tool.Tool().GetServiceId())
	assert.Equal(t, "mcp_config_error", tool.MCPTool().Name)

	resp, err := tool.Execute(context.Background(), nil)
	assert.NoError(t, err)

	respMap, ok := resp.(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, errMsg, respMap["error"])
	assert.Equal(t, "Configuration Failed", respMap["status"])
}
