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

func TestLocalCommandTool_Execute_JSONProtocol_StderrCapture(t *testing.T) {
	tool := &v1.Tool{
		Name:        proto.String("test-tool-json-stderr"),
		Description: proto.String("A test tool that fails"),
	}
	// Command that writes to stderr and exits with error, producing invalid JSON (empty stdout)
	service := &configv1.CommandLineUpstreamService{
		Command:               proto.String("sh"),
		Local:                 proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "echo 'something went wrong' >&2; exit 1"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:   "test-tool-json-stderr",
		Arguments:  map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)

	// This assertion should fail before fix
	assert.Contains(t, err.Error(), "something went wrong")
}
