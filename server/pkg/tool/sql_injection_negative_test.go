package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSQLInjection_Negative(t *testing.T) {
	// Defense in Depth: SQL Injection via shell injection checks.
	// Even if checkSQLInjection is skipped for quoted strings,
	// checkForShellInjection should block the quote character itself, preventing the SQL breakout.

	t.Run("sql_injection_quoted_bypass_attempt", func(t *testing.T) {
		cmd := "psql"
		// Template uses input in single quotes: SELECT * FROM users WHERE id = '{{id}}'
		// This is a common pattern.
		args := []string{"-c", "SELECT * FROM users WHERE id = '{{id}}'"}

		toolDef := v1.Tool_builder{
			Name: proto.String("sql-tool"),
		}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: proto.String(cmd),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: args,
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("id"),
					}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		// Attack: 1' OR '1'='1
		// If injected: SELECT * FROM users WHERE id = '1' OR '1'='1'
		req := &ExecutionRequest{
			ToolName:   "test",
			ToolInputs: []byte(`{"id": "1' OR '1'='1"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// Expect error because ' is blocked by checkSingleQuotedInjection
		assert.Error(t, err, "Should detect SQL injection attempt via quote breakout")
		if err != nil {
			assert.Contains(t, err.Error(), "shell injection detected", "Generic shell injection check should catch the quote")
		}
	})
}
