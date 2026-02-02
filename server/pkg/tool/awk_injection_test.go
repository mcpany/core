package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_AwkInjection_PipeSh(t *testing.T) {
	t.Parallel()

	// Define a tool that uses 'awk'.
	// We pass a user-provided script inside single quotes.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"script": structpb.NewStructValue(&structpb.Struct{
                        Fields: map[string]*structpb.Value{
                            "type": structpb.NewStringValue("string"),
                        },
                    }),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("awk-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: Use awk's pipe to execute shell commands
	// 'BEGIN { print "pwned" | "sh" }'
	// This does not use system(), so it bypasses the simple "system(" check.
	payload := `BEGIN { print "echo pwned" | "sh" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Let's use DryRun = true. It should return "dry_run": true if validation passes.
	// If validation fails (as we hope), it returns error.
	req.DryRun = true

	_, err := localTool.Execute(context.Background(), req)

	// We expect an error "awk injection detected" or "shell injection detected"
	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk injection with pipe passed validation")
	} else {
		t.Logf("Validation blocked payload: %v", err)
	}

    // Assert Error to confirm the fix is working
    assert.Error(t, err, "Expected validation error for awk injection")
    assert.Contains(t, err.Error(), "awk injection detected")
}

func TestLocalCommandTool_Awk_Benign_LogicalOr(t *testing.T) {
	t.Parallel()

	// Define a tool that uses 'awk'.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"script": structpb.NewStructValue(&structpb.Struct{
                        Fields: map[string]*structpb.Value{
                            "type": structpb.NewStringValue("string"),
                        },
                    }),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("awk-tool-benign"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: Use awk's logical OR '||' which is safe and necessary.
	// 'BEGIN { if (1 || 0) print "ok" }'
	payload := `BEGIN { if (1 || 0) print "ok" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool-benign",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)
	req.DryRun = true

	_, err := localTool.Execute(context.Background(), req)

	// We expect success (nil error)
	assert.Nil(t, err, "Benign logical OR should be allowed")
}
