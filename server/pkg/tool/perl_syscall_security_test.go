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

func TestLocalCommandTool_Perl_Syscall_RCE(t *testing.T) {
	// This test demonstrates RCE vulnerability in Perl using syscall
	tool := v1.Tool_builder{Name: proto.String("test-tool-perl-syscall")}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
			}.Build(),
		},
		Args: []string{"-e", "{{code}}"},
	}.Build()

	// Policies and ID are needed for NewLocalCommandTool
	policies := []*configv1.CallPolicy{}

	localTool := NewLocalCommandTool(tool, service, callDef, policies, "call-id")

	// Attempt RCE using syscall which bypasses current checks
	// syscall 59 is execve on Linux x86_64
	// usage: syscall(SYS_execve, path, argv, envp)
	// We use q// for strings to avoid quotes which are blocked
	// We use comma , which is allowed
	// We use space which is allowed for interpreters (not shells)

	// Payload: syscall 59, q/bin/echo/, 0, 0
	// This should execute /bin/echo if successful.
	// Note: 59 is arch specific, but for validation logic test, it doesn't matter if it actually works on the OS,
	// what matters is if the VALIDATION blocks it.

	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-perl-syscall",
		Arguments: map[string]interface{}{
			"code": "syscall 59, q/bin/echo/, 0, 0",
		},
	}
	inputs, _ := json.Marshal(reqAttack.Arguments)
	reqAttack.ToolInputs = inputs

	_, err := localTool.Execute(context.Background(), reqAttack)

	// If err contains "shell injection detected" or "interpreter injection detected", it means validation failed -> Secure.
	if err != nil {
		if assert.Contains(t, err.Error(), "injection detected") {
             // Validation worked.
			 t.Log("Validation correctly blocked syscall injection")
        } else {
             t.Logf("Validation passed (vulnerable), but execution failed: %v", err)
			 // Since we are testing validation logic, failure to execute (e.g. exit code) means validation passed.
             t.Fail()
        }
	} else {
		// Validation passed (vulnerable)
		t.Log("Perl syscall payload was NOT blocked! (Execution succeeded)")
		t.Fail()
	}
}
