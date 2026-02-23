// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RubySyscallInjection(t *testing.T) {
	// This test demonstrates that Ruby syscall should be BLOCKED in unquoted context.

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("ruby-syscall"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// Ruby script: eval(input)
	// We use -e to execute code directly.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: puts syscall 20 (getpid)
	// This should be blocked.
	payload := "puts syscall 20"

	req := &ExecutionRequest{
		ToolName: "ruby-syscall",
		Arguments: map[string]interface{}{
			"code": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// After fix, this MUST fail.
	assert.Error(t, err, "Expected execution to be blocked")
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "interpreter injection detected"), "Expected interpreter injection error, got: %v", err)
	}
}
