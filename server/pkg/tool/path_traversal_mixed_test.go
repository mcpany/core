// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_PathTraversal_MixedEncoded(t *testing.T) {
	t.Parallel()

	toolDef := mcp_router_v1.Tool_builder{
		Name: proto.String("test-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("arg"),
				}.Build(),
			}.Build(),
		},
		Args: []string{"{{arg}}"},
	}.Build()
	localTool := tool.NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	testCases := []string{
		"%2e.",
		".%2e",
		"%2E.",
		".%2E",
		"%2e%2E",
		"%2E%2e",
	}

	for _, payload := range testCases {
		t.Run(payload, func(t *testing.T) {
			inputStr := payload + "/secret"
			inputs := json.RawMessage(`{"arg": "` + inputStr + `"}`)
			req := &tool.ExecutionRequest{ToolInputs: inputs}
			_, err := localTool.Execute(context.Background(), req)

			// We expect the security check to catch this.
			// If the vulnerability exists, this will FAIL (err will be nil).
			require.Error(t, err, "Expected path traversal error for %q", payload)
			assert.Contains(t, err.Error(), "path traversal attempt detected")
		})
	}
}
