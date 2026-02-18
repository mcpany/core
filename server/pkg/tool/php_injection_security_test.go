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

func TestPHPInjection(t *testing.T) {
	// Sentinel Security Update: Reproduction test case for PHP injection bypass.
	// This test simulates a PHP command execution where the user controls the code.
	// We check if dangerous functions like 'passthru' are blocked.

	tool := v1.Tool_builder{
		Name:                proto.String("php_exec"),
		DisplayName:         proto.String("PHP Exec"),
		Description:         proto.String("Executes PHP code"),
		UnderlyingMethodFqn: proto.String("php_exec"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
		// Args are in CallDefinition
	}.Build()

	// We use a template that wraps input in eval and single quotes.
	// Template: eval('{{code}}');
	// QuoteLevel = 2 (Single).
	// This blocks single quotes (breakout), but allows double quotes and parens.
	// Since it's inside eval(), the content IS executed as PHP code.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "eval('{{code}}');"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolInstance := NewLocalCommandTool(tool, service, callDef, nil, "php_exec")

	ctx := context.Background()

	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "Safe PHP code",
			input:     "echo \"hello\";",
			shouldErr: false,
		},
		{
			name:      "Blocked system()",
			input:     "system(\"ls\");",
			shouldErr: true,
		},
		{
			name:      "Blocked exec()",
			input:     "exec(\"ls\");",
			shouldErr: true,
		},
		{
			name:      "Bypass attempt with passthru()",
			input:     "passthru(\"ls\");",
			shouldErr: true, // Should be blocked, but currently passes (vulnerability)
		},
		{
			name:      "Bypass attempt with shell_exec()",
			input:     "shell_exec(\"ls\");",
			shouldErr: true, // Should be blocked
		},
		{
			name:      "Bypass attempt with assert()",
			input:     "assert(\"system(\\\"ls\\\")\");", // Use double quotes, escaped
			shouldErr: true, // Should be blocked
		},
		{
			name:      "Bypass attempt with proc_open()",
			input:     "proc_open(\"ls\", [], $pipes);",
			shouldErr: true, // Should be blocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputMap := map[string]interface{}{"code": tt.input}
			toolInputs, err := json.Marshal(inputMap)
			assert.NoError(t, err)

			req := &ExecutionRequest{
				ToolName:   "php_exec",
				ToolInputs: toolInputs,
				DryRun:     true,
			}

			_, err = toolInstance.Execute(ctx, req)
			if tt.shouldErr {
				if err == nil {
					t.Logf("Expected error for input %q, but got none (VULNERABLE)", tt.input)
					t.Fail()
				} else {
                    t.Logf("Got expected error: %v", err)
                }
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
			}
		})
	}
}
