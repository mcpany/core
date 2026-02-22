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

func TestCurlSSRFBypass(t *testing.T) {
	// Save original IsSafeURL
	originalIsSafeURL := validation.IsSafeURL
	defer func() {
		validation.IsSafeURL = originalIsSafeURL
	}()

	// Mock IsSafeURL to fail for our target host
	validation.IsSafeURL = func(urlStr string) error {
		if strings.Contains(urlStr, "localtest.me") {
			return fmt.Errorf("host \"localtest.me\" resolves to unsafe IP 127.0.0.1: loopback address is not allowed")
		}
		// Simulate NXDOMAIN or invalid host for non-host arguments
		if strings.Contains(strings.ToLower(urlStr), "header") || strings.Contains(urlStr, "-v") {
			return fmt.Errorf("no IP addresses found for host")
		}

		// For http://127.0.0.1 (used in Control case), we should fail too if we want to mimic IsSafeURL
		if strings.Contains(urlStr, "127.0.0.1") {
			return fmt.Errorf("loopback address is not allowed")
		}

		return nil
	}

	// Define a curl tool using builders
	toolDef := v1.Tool_builder{
		Name: proto.String("curl"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
	}.Build()

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

	// Create tool
	tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test-call")

	t.Run("Block schema-less loopback hostname", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName:   "curl",
			ToolInputs: []byte(`{"url": "localtest.me"}`),
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "unsafe network argument")
			assert.Contains(t, err.Error(), "not allowed")
		}
	})

	t.Run("Block explicit loopback URL", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName:   "curl",
			ToolInputs: []byte(`{"url": "http://127.0.0.1"}`),
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "unsafe url argument")
		}
	})

	t.Run("Allow headers (usability)", func(t *testing.T) {
		// This simulates a header argument passed to curl
		// Although typically passed via Args template like "-H", "{{header}}"
		// Here we assume {{url}} is replaced by header string to test validation logic
		req := &ExecutionRequest{
			ToolName:   "curl",
			ToolInputs: []byte(`{"url": "X-Custom-Header: value"}`),
		}
		_, err := tool.Execute(context.Background(), req)
		// We expect execution failure (curl fails), but NOT validation failure
		if err != nil {
			assert.NotContains(t, err.Error(), "unsafe network argument", "Should not block non-resolvable headers")
		}
	})

    // Note: Flags starting with - are blocked by checkForArgumentInjection anyway,
    // so we don't test them here as "Allowed", but we could test that they don't fail with "unsafe network argument"
    // but rather "argument injection".
}
