/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMCPTool_Execute_Bug_OriginalName(t *testing.T) {
	// This test demonstrates the fix where MCPTool uses the original name (from tool definition)
	// instead of the sanitized name (from request parsing) when calling the upstream service.

	originalName := "my.tool"
	// In util.go: nonWordChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	// "." matches this regex, so it is replaced by "".
	// So "my.tool" -> "mytool" and since length changed, hash is appended.

	sanitizedName, err := util.SanitizeToolName(originalName)
	require.NoError(t, err)

	serviceID := "myservice"
	// The tool is registered as "myservice.sanitizedName".

	// Upstream expects "my.tool".

	mockClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			// This assertion confirms the fix.
			// It should send "my.tool", not "mytool_...".
			if params.Name != originalName {
				return nil, errors.New("wrong tool name: " + params.Name)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: `{"output":"result"}`},
				},
			}, nil
		},
	}

	toolProto := &v1.Tool{}
	toolProto.SetName(originalName)   // "my.tool"
	toolProto.SetServiceId(serviceID) // "myservice"
	mcpTool := tool.NewMCPTool(toolProto, mockClient, &configv1.MCPCallDefinition{})

	inputs := json.RawMessage(`{}`)
	// The request comes in with the registered name
	req := &tool.ExecutionRequest{
		ToolName:   serviceID + "." + sanitizedName,
		ToolInputs: inputs,
	}

	_, err = mcpTool.Execute(context.Background(), req)

	// We assert that it succeeds.
	require.NoError(t, err)
}
