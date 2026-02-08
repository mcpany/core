package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_SSRF_Prevention(t *testing.T) {
	// This test verifies that command arguments that look like URLs are validated against SSRF.

	// Unset the env var that TestMain sets to allow all IPs
	originalEnv := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	defer os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", originalEnv)

	// Mock validation.IsSafeURL to fail for unsafe URLs (for http/https).
	// TestMain disables it globally, so we must enable it here to test the invocation.
	originalIsSafeURL := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	validation.IsSafeURL = func(urlStr string) error {
		if strings.Contains(urlStr, "127.0.0.1") || strings.Contains(urlStr, "localhost") {
			return fmt.Errorf("loopback address is not allowed (mock)")
		}
		if strings.Contains(urlStr, "169.254.169.254") {
			return fmt.Errorf("link-local address is not allowed (mock)")
		}
		return nil
	}

	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-curl"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build()}.Build(),
		},
		Args: []string{"{{url}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Case 1: Unsafe Internal IP (Loopback) via HTTP
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-curl",
		Arguments: map[string]interface{}{
			"url": "http://127.0.0.1/admin",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	_, err := localTool.Execute(context.Background(), reqAttack)

	assert.Error(t, err, "Should block loopback IP")
	if err != nil {
		assert.Contains(t, err.Error(), "loopback address is not allowed")
	}

	// Case 2: Unsafe Cloud Metadata IP via HTTP
	reqMetadata := &ExecutionRequest{
		ToolName: "test-tool-curl",
		Arguments: map[string]interface{}{
			"url": "http://169.254.169.254/latest/meta-data/",
		},
	}
	reqMetadata.ToolInputs, _ = json.Marshal(reqMetadata.Arguments)

	_, err = localTool.Execute(context.Background(), reqMetadata)
	assert.Error(t, err, "Should block metadata IP")
	if err != nil {
		assert.Contains(t, err.Error(), "link-local address is not allowed")
	}

	// Case 3: Argument is not a URL (should pass SSRF check)
	reqNormal := &ExecutionRequest{
		ToolName: "test-tool-curl",
		Arguments: map[string]interface{}{
			"url": "filename.txt",
		},
	}
	reqNormal.ToolInputs, _ = json.Marshal(reqNormal.Arguments)

	_, err = localTool.Execute(context.Background(), reqNormal)
	if err != nil {
		assert.NotContains(t, err.Error(), "loopback address is not allowed")
		assert.NotContains(t, err.Error(), "link-local address is not allowed")
	}

	// Case 4: Dangerous Scheme (gopher)
	reqGopher := &ExecutionRequest{
		ToolName: "test-tool-curl",
		Arguments: map[string]interface{}{
			"url": "gopher://127.0.0.1/1",
		},
	}
	reqGopher.ToolInputs, _ = json.Marshal(reqGopher.Arguments)
	_, err = localTool.Execute(context.Background(), reqGopher)
	assert.Error(t, err, "Should block gopher scheme")
	if err != nil {
		assert.Contains(t, err.Error(), "scheme \"gopher\" is not allowed")
	}

	// Case 5: SSH with Loopback (Manual Check)
	reqSSH := &ExecutionRequest{
		ToolName: "test-tool-curl",
		Arguments: map[string]interface{}{
			"url": "ssh://127.0.0.1/repo",
		},
	}
	reqSSH.ToolInputs, _ = json.Marshal(reqSSH.Arguments)
	_, err = localTool.Execute(context.Background(), reqSSH)
	assert.Error(t, err, "Should block ssh to loopback")
	if err != nil {
		assert.Contains(t, err.Error(), "loopback address is not allowed")
	}
}
