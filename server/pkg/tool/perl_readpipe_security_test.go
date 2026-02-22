// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestPerlReadpipeSecurity attempts to bypass the protection using Perl's readpipe function.
func TestPerlReadpipeSecurity(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		// Use -e {{code}}
        Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{
		Name: proto.String("perl_eval"),
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "call_id")

	// Payload: Use readpipe with qw list operator to avoid quotes and parens.
    // readpipe executes the command and returns output.
    // qw/echo RCE_SUCCESS/ creates a string "echo RCE_SUCCESS" (well, list of words but Perl joins/concats? No readpipe takes scalar EXPR).
    // qw/str/ in scalar context returns the string? No, qw returns a list.
    // In scalar context, qw returns the last element?
    // Let's test scalar(qw/echo RCE_SUCCESS/). -> RCE_SUCCESS.
    // readpipe takes EXPR.
    // So readpipe qw/echo RCE_SUCCESS/ executes "RCE_SUCCESS"? No.
    // Wait. qw/echo RCE_SUCCESS/ is ('echo', 'RCE_SUCCESS').
    // In scalar context, comma operator returns last element? No.
    // qw returns a list. List in scalar context returns last element in Perl 5.10+?
    // Actually, readpipe takes scalar expression.
    // q/echo RCE_SUCCESS/ returns the string!
    // q// is single quote operator.
    // q/echo RCE_SUCCESS/ -> 'echo RCE_SUCCESS'.
    // checkUnquotedInjection blocks quotes? No, it blocks ' and ".
    // It does NOT block /.
    // It does NOT block q.

    // So payload: print readpipe q/echo RCE_SUCCESS/
	payload := `print readpipe q/echo RCE_SUCCESS/`

	req := &ExecutionRequest{
		ToolName:   "perl_eval",
		ToolInputs: []byte(`{"code": "` + payload + `"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		if strings.Contains(err.Error(), "interpreter injection detected") ||
		   strings.Contains(err.Error(), "shell injection detected") {
			t.Logf("Secure: Injection was blocked: %v", err)
			return
		}
		t.Logf("Execution failed: %v", err)
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)

		if strings.Contains(stdout, "RCE_SUCCESS") {
			t.Log("VULNERABLE: Perl readpipe injection succeeded!")
            t.Fatal("Vulnerability confirmed: Perl readpipe injection succeeded")
		} else {
            t.Logf("Executed but no RCE output? Stdout: %q", stdout)
        }
	}
}
