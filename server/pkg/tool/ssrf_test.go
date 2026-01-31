package tool

import (
	"context"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_SSRF(t *testing.T) {
	// Override IsSafeURL for this test
	originalIsSafeURL := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	// Mock IsSafeURL to behave somewhat like the real one
	validation.IsSafeURL = func(urlStr string) error {
		if strings.HasPrefix(urlStr, "ftp://") {
			return fmt.Errorf("unsupported scheme: ftp")
		}
		if urlStr == "http://internal.local" {
			return fmt.Errorf("unsafe url: loopback address is not allowed")
		}
		return nil
	}

	// Configure a simple tool (echo)
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("echo"),
		WorkingDirectory: proto.String("."),
	}.Build()

	// Define a parameter 'url'
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("echo_tool")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"echo_tool",
	)

	// Test Case 1: Internal URL (Should Fail)
	req := &ExecutionRequest{
		ToolName:   "echo_tool",
		ToolInputs: []byte(`{"url": "http://internal.local"}`),
	}
	_, err := tool.Execute(context.Background(), req)

	// We expect an error
	assert.Error(t, err, "Expected error for unsafe URL")
	if err != nil {
		assert.Contains(t, err.Error(), "unsafe url")
	}

	// Test Case 2: External URL (Should Pass)
	req2 := &ExecutionRequest{
		ToolName:   "echo_tool",
		ToolInputs: []byte(`{"url": "http://external.com"}`),
	}
	res, err := tool.Execute(context.Background(), req2)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Test Case 3: FTP URL (Should Fail)
	req3 := &ExecutionRequest{
		ToolName:   "echo_tool",
		ToolInputs: []byte(`{"url": "ftp://example.com"}`),
	}
	_, err = tool.Execute(context.Background(), req3)
	assert.Error(t, err, "Expected error for FTP URL")
	if err != nil {
		assert.Contains(t, err.Error(), "ssrf")
		assert.Contains(t, err.Error(), "unsupported scheme")
	}
}
