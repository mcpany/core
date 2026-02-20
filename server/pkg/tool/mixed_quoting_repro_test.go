package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMixedQuotingInjection(t *testing.T) {
	// Vulnerability Reproduction: Mixed Quoting Context Injection
	// If a placeholder is used in both double quotes (Level 1) and single quotes (Level 2),
	// the current implementation picks the minimum level (Level 1).
	// Level 1 checks forbid `"` `$` `\` etc., but ALLOW `'`.
	// Level 2 checks forbid `'`.
	// By picking Level 1, we allow `'` to pass through.
	// This `'` then breaks out of the single-quoted context in the second usage.

	t.Run("mixed_quoting_bypass", func(t *testing.T) {
		cmd := "bash"
		// Template uses input in double quotes AND single quotes
		// echo "{{input}}" '{{input}}'
		args := []string{"-c", "echo \"{{input}}\" '{{input}}'"}

		toolDef := v1.Tool_builder{
			Name: proto.String("test-tool"),
		}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: proto.String(cmd),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: args,
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("input"),
					}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		// We inject a single quote.
		// Safe in double quotes: "foo'bar" -> foo'bar
		// Dangerous in single quotes: 'foo'bar' -> foo (end string) bar (unquoted) (start string)
		// We append a command to verify injection: ' ; true ; '
		req := &ExecutionRequest{
			ToolName:   "test",
			ToolInputs: []byte(`{"input": "' ; true ; '"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// In the fixed version, this MUST return an error.
		assert.Error(t, err, "Should detect injection in mixed quoting context")
		if err != nil {
			assert.Contains(t, err.Error(), "shell injection detected")
		}
	})
}
