// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPythonUnicodeBypass(t *testing.T) {
	// Vulnerability: Python 3 normalizes identifiers (NFKC), allowing fullwidth characters
	// to execute as standard ASCII keywords. Security checks that only look at ASCII or
	// rely on simple lowercase conversion without normalization can be bypassed.
	t.Run("Python_Unicode_Keyword_Bypass", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("python_unicode_tool"),
		}).Build()
		cmd := "python3"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Template: eval('{{code}}')
		// This simulates a tool that dangerously evals input, which the security layer
		// is intended to protect against by blocking 'eval', 'system', etc.
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "eval('{{code}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Input: __ｉｍｐｏｒｔ__("os").ｓｙｓｔｅｍ("echo injected")
		// Using Fullwidth Latin Small Letters (U+FF49, U+FF4D, etc.)
		// These normalize to "import", "os", "system".
		// We use double quotes for strings to avoid breaking out of the single-quoted template context.
		input := "__ｉｍｐｏｒｔ__(\"ｏｓ\").ｓｙｓｔｅｍ(\"echo injected\")"

		req := &ExecutionRequest{
			ToolName: "python_unicode_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		// Assert that the security layer detects this injection attempt.
		// If the vulnerability is present, err will be nil (or an execution error unrelated to security),
		// and this assertion will fail.
		if assert.Error(t, err, "Should detect injection even with unicode obfuscation") {
			assert.Contains(t, err.Error(), "interpreter injection detected", "Should detect interpreter injection keywords")
		}
	})
}
