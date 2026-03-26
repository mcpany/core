// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RubySyscallInjection(t *testing.T) {
	// This test demonstrates RCE vulnerability in Ruby via syscall.
	// We use 'ruby -e' with unquoted argument to bypass quote checks.
	// Since Ruby is not a shell, spaces are allowed.

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("ruby-syscall"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: syscall 20 (getpid)
	// We avoid quotes to pass checkUnquotedInjection (which blocks ' and ").
	// syscall is not in the blocked keywords list.
	payload := "syscall 20"

	req := &ExecutionRequest{
		ToolName: "ruby-syscall",
		Arguments: map[string]interface{}{
			"code": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect this to execute successfully (VULNERABLE)
	// Or fail with an error if fixed.

	_, err := localTool.Execute(context.Background(), req)

	// If vulnerable, err might be nil (execution success) or execution error (but passed security check)
	// If fixed, err should contain "interpreter injection detected"

	if err == nil {
		t.Fatal("VULNERABILITY CONFIRMED: Ruby syscall injection passed security checks (err is nil)")
	}

	// If it fails, check if it's the security check
	assert.ErrorContains(t, err, "interpreter injection detected", "VULNERABILITY CONFIRMED: Error did not indicate security block")
}
