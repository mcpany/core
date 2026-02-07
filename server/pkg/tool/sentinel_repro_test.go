// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSentinel_ShellInjection_InterpreterBypass(t *testing.T) {
	// This test attempts to bypass the shell injection protection by nesting an interpreter
	// (Ruby) inside a shell command (sh -c).
	// The injection uses Ruby string interpolation #{} which is valid inside shell single quotes
	// but dangerous when executed by Ruby.

	t.Parallel()

	tmpDir := t.TempDir()
	proofFile := tmpDir + "/sentinel_rce_proof_repro"

	tool := v1.Tool_builder{
		Name: proto.String("vulnerable-tool"),
	}.Build()

	// Tool definition: sh -c "ruby -e 'print \"{{msg}}\"'"
	// Arg 1: -c
	// Arg 2: ruby -e 'print "{{msg}}"'
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "ruby -e 'print \"{{msg}}\"'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Injection payload: #{system('touch /tmp/sentinel_rce_proof_repro')}
	// We use Kernel.send(:system, ...) to bypass the check for "system(" string.
	// We use double quotes inside the string to avoid single-quote check.
	payload := "#{Kernel.send(:system, \"touch " + proofFile + "\")}"

	req := &ExecutionRequest{
		ToolName: "vulnerable-tool",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// Check if file exists
	if _, statErr := os.Stat(proofFile); statErr == nil {
		t.Errorf("Security Vulnerability Confirmed: RCE payload executed, proof file created at %s", proofFile)
	} else {
		// If file doesn't exist, we are good.
		// We expect an error if the fix is implemented.
		t.Logf("Proof file not created. Execution error: %v", err)
		if err == nil {
			t.Log("Warning: Execution succeeded but file not created? Maybe ruby failed or payload incorrect?")
		} else {
			assert.Contains(t, err.Error(), "shell injection detected", "Expected shell injection detection error")
		}
	}
}

func TestSentinel_ShellInjection_NestedShell_CommandSubstitution(t *testing.T) {
	// This test attempts to bypass the shell injection protection by using command substitution $()
	// inside a nested shell command.
	// Single quotes in the outer shell do NOT expand $(), but if passed to an inner shell, it will be expanded.

	t.Parallel()

	tmpDir := t.TempDir()
	proofFile := tmpDir + "/sentinel_rce_proof_nested"

	tool := v1.Tool_builder{
		Name: proto.String("vulnerable-tool-nested"),
	}.Build()

	// Tool definition: sh -c 'echo {{msg}}'
	// Outer shell: sh
	// Arg 1: -c
	// Arg 2: echo '{{msg}}' (Note: usually templates puts quotes around placeholders)
	// Let's assume the template is: echo '{{msg}}'
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "sh -c 'touch {{msg}}'"}, // Nested sh for clarity
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Injection payload: /tmp/sentinel_rce_proof_nested; $(touch /tmp/sentinel_rce_proof_nested)
	// Actually, if the template is: sh -c 'touch {{msg}}'
	// We want {{msg}} to be: /tmp/foo; touch /tmp/sentinel_rce_proof_nested
	// But ; is blocked in Single Quotes (Level 2).
	// But $(...) is NOT blocked.
	// So we can use: $(touch /tmp/sentinel_rce_proof_nested)
	// Result: sh -c 'touch $(touch /tmp/sentinel_rce_proof_nested)'
	// The inner shell executes $(touch ...), which creates the file. The output (empty) is passed to touch.
	// touch "" -> fails? or touches current dir?
	// touch /tmp/sentinel_rce_proof_nested returns nothing.
	// So inner shell runs: touch ""
	// But the SIDE EFFECT (creation) happens during expansion.

	payload := "$(" + "touch " + proofFile + ")"

	req := &ExecutionRequest{
		ToolName: "vulnerable-tool-nested",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// Check if file exists
	if _, statErr := os.Stat(proofFile); statErr == nil {
		t.Errorf("Security Vulnerability Confirmed: RCE payload executed, proof file created at %s", proofFile)
	} else {
		t.Logf("Proof file not created. Execution error: %v", err)
		if err != nil {
			assert.Contains(t, err.Error(), "shell injection detected", "Expected shell injection detection error")
		}
	}
}
