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

func TestSecurity_UsabilityBypass(t *testing.T) {
	// Setup a Python tool
	pythonService := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python"),
	}).Build()

	pythonCallDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{
					Name: proto.String("msg"),
				}).Build(),
			}).Build(),
		},
	}).Build()

	toolV1 := (&v1.Tool_builder{
		Name: proto.String("test"),
	}).Build()

	pythonTool := NewLocalCommandTool(
		toolV1,
		pythonService,
		pythonCallDef,
		nil,
		"call1",
	)

	// Setup a Ruby tool
	rubyService := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
	}).Build()

	rubyCallDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "puts '{{msg}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{
					Name: proto.String("msg"),
				}).Build(),
			}).Build(),
		},
	}).Build()

	rubyTool := NewLocalCommandTool(
		toolV1,
		rubyService,
		rubyCallDef,
		nil,
		"call2",
	)

	tests := []struct {
		name      string
		tool      Tool
		input     string
		shouldFail bool
		errorMsg  string
	}{
		// Python - Safe cases (Verified in previous test, but good to have)
		{
			name:      "Python safe word 'system'",
			tool:      pythonTool,
			input:     "The system is down",
			shouldFail: false,
		},
		{
			name:      "Python safe word 'os'",
			tool:      pythonTool,
			input:     "I use os x",
			shouldFail: false,
		},

		// Python - Unsafe cases (Must block)
		{
			name:      "Python import statement",
			tool:      pythonTool,
			input:     "import os",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"import\"",
		},
		{
			name:      "Python os.system",
			tool:      pythonTool,
			input:     "os.system('ls')",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"os\" followed by '.'",
		},
		{
			name:      "Python os[system]",
			tool:      pythonTool,
			input:     "os['system']('ls')",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"os\" followed by '['",
		},
		{
			name:      "Python eval call",
			tool:      pythonTool,
			input:     "eval('1+1')",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"eval\" followed by '('",
		},
		{
			name:      "Python __import__",
			tool:      pythonTool,
			input:     "__import__('os')",
			shouldFail: true,
			errorMsg:  "value contains '__import__'",
		},
		{
			name:      "Python spaced call",
			tool:      pythonTool,
			input:     "eval ('1+1')",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"eval\" followed by '('",
		},
		// Python quoted safe usage
		{
			name:      "Python quoted safe usage",
			tool:      pythonTool,
			input:     `I said "os.system" is dangerous`,
			shouldFail: false, // Should pass because it is inside quotes (double quotes inside single quoted template is safe from shell injection, and now safe from interpreter injection)
		},

		// Ruby - Strict mode checks
		{
			name:      "Ruby system call unquoted",
			tool:      rubyTool,
			input:     "system 'ls'",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"system\"",
		},
		{
			name:      "Ruby system call quoted",
			tool:      rubyTool,
			input:     "system('ls')",
			shouldFail: true,
			errorMsg:  "dangerous keyword \"system\"",
		},
		{
			name:      "Ruby safe word 'system' (Strict mode blocks it anyway)",
			tool:      rubyTool,
			input:     "The system is down",
			shouldFail: true, // Ruby is strict, so it blocks "system" even if not obviously a call
			errorMsg:  "dangerous keyword \"system\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inputMap := map[string]string{"msg": tc.input}
			inputJson, _ := json.Marshal(inputMap)

			req := &ExecutionRequest{
				ToolName: "test",
				ToolInputs: inputJson,
				DryRun: true,
			}

			_, err := tc.tool.Execute(context.Background(), req)

			if tc.shouldFail {
				if assert.Error(t, err, "Input should be blocked: %s", tc.input) {
					// We verify that the error is indeed related to the injection, not something else
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Input should be allowed: %s", tc.input)
			}
		})
	}
}
