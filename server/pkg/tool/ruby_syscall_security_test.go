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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RubySyscallInjection(t *testing.T) {
	// This test demonstrates that Ruby `syscall` execution is possible,
	// allowing arbitrary system calls (including execve) and bypassing restrictions.

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("ruby-syscall-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "puts {{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: syscall 20 (getpid)
	// We use `syscall 20` which returns the PID.
	// This proves code execution via syscall without side effects.
	// Since `syscall` is not blocked, this should succeed before the fix.
	payload := "syscall 20"

	req := &ExecutionRequest{
		ToolName: "ruby-syscall-tool",
		Arguments: map[string]interface{}{
			"code": payload,
		},
	}
	var err error
	req.ToolInputs, err = json.Marshal(req.Arguments)
	require.NoError(t, err)

	result, err := localTool.Execute(context.Background(), req)

	// Before fix: This should succeed (err == nil) because syscall is not blocked.
	// After fix: This should fail with "interpreter injection detected".

	if err != nil {
		if strings.Contains(err.Error(), "interpreter injection detected") {
			t.Log("PASS: Ruby syscall blocked correctly")
			return
		}
		t.Logf("Blocked by other error: %v", err)
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, _ := resMap["stdout"].(string)
		t.Logf("Stdout: %s", stdout)
		assert.Fail(t, "VULNERABILITY CONFIRMED: Ruby syscall execution successful (not blocked)")
	}
}
