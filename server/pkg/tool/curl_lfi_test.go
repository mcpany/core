package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSentinelExploit_CurlLFI(t *testing.T) {
	t.Parallel()

	// Create a dummy file that simulates a sensitive file (e.g., /etc/passwd)
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err := os.WriteFile(secretFile, []byte("sensitive_data"), 0600)
	assert.NoError(t, err)

	// Define the tool
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"data": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("curl-tool"),
		InputSchema: inputSchema,
	}.Build()

	// Use 'curl' as the command since it's in the isShellCommand list and supports @file
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-d", "{{data}}", "http://localhost"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("data")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "exploit-id")

	// Payload: @/path/to/secret.txt
	// This exploits the fact that filepath.IsAbs("@/...") is false, but curl treats it as a file path.
	payload := "@" + secretFile

	req := &ExecutionRequest{
		ToolName: "curl-tool",
		Arguments: map[string]interface{}{
			"data": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err = localTool.Execute(context.Background(), req)

	// With the fix, we expect an error catching the absolute path inside @file
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "absolute path detected in @file argument")
}

func TestSentinelExploit_CurlTraversal(t *testing.T) {
	t.Parallel()

	// Define the tool
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"data": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("curl-tool-traversal"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-d", "{{data}}", "http://localhost"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("data")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "exploit-id-traversal")

	// Payload: @../secret.txt
	// This exploits path traversal bypass if @ prefix is not handled
	payload := "@../secret.txt"

	req := &ExecutionRequest{
		ToolName: "curl-tool-traversal",
		Arguments: map[string]interface{}{
			"data": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	// With the fix, we expect an error catching the traversal inside @file
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected in @file argument")
}

func TestSentinelExploit_SafeHandle(t *testing.T) {
	t.Parallel()

	// Define the tool
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"data": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("echo-handle"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{data}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("data")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "safe-handle-id")

	// Payload: @myhandle
	// This should be allowed as it is neither absolute path nor traversal
	payload := "@myhandle"

	req := &ExecutionRequest{
		ToolName: "echo-handle",
		Arguments: map[string]interface{}{
			"data": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	// Should pass validation
	assert.NoError(t, err)
}
