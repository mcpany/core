// Copyright 2026 Author(s) of MCP Any
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

func TestFindInjection(t *testing.T) {
	t.Run("Find_Exec_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("find_tool"),
		}).Build()
		cmd := "find"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{".", "-name", "{{filename}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("filename"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using -exec with +
		// We avoid {}, using a command that accepts arguments (like echo or ls)
		input := "foo -exec echo pwned +"

		req := &ExecutionRequest{
			ToolName: "find_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"filename": %q}`, input)),
			Arguments: map[string]interface{}{
				"filename": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "find injection detected", "Should detect -exec injection")
	})

	t.Run("Find_Delete_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("find_delete_tool"),
		}).Build()
		cmd := "find"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{".", "-name", "{{filename}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("filename"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using -delete
		input := "foo -delete"

		req := &ExecutionRequest{
			ToolName: "find_delete_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"filename": %q}`, input)),
			Arguments: map[string]interface{}{
				"filename": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "find injection detected", "Should detect -delete injection")
	})
}

func TestCheckFindInjection(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		base    string
		wantErr bool
	}{
		{
			name:    "safe find command with spaces",
			val:     ". -name \"*.go\" -type f",
			base:    "find",
			wantErr: false,
		},
		{
			name:    "not find base command",
			val:     "-exec rm -rf /",
			base:    "grep",
			wantErr: false,
		},
		{
			name:    "malicious exec flag",
			val:     ". -name \"*.tmp\" -exec rm {} \\;",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "malicious execdir flag",
			val:     ". -name \"*.tmp\" -execdir rm {} \\;",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "malicious ok flag",
			val:     ". -name \"*.tmp\" -ok rm {} \\;",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "malicious okdir flag",
			val:     ". -name \"*.tmp\" -okdir rm {} \\;",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "malicious delete flag",
			val:     ". -name \"*.tmp\" -delete",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "malicious exec flag uppercase",
			val:     ". -name \"*.tmp\" -EXEC rm {} \\;",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "malicious flag embedded inside safe string",
			val:     ". -name \"-exec-is-bad\"",
			base:    "find",
			wantErr: false,
		},
		{
			name:    "malicious flag with prefix spacing",
			val:     ". -name \"*.tmp\"    -exec   rm {} \\;",
			base:    "find",
			wantErr: true,
		},
		{
			name:    "empty value",
			val:     "",
			base:    "find",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFindInjection(tt.val, tt.base); (err != nil) != tt.wantErr {
				t.Errorf("checkFindInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
