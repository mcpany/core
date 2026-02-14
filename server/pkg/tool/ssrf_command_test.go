// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"fmt"
	"os"
	"strings"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_SSRF_Vulnerability(t *testing.T) {
	// This test demonstrates that LocalCommandTool protects against SSRF
	// when passing URL arguments.

	// Save and restore environment/variables
	originalEnv := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "") // Disable bypass
	defer os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", originalEnv)

	originalIsSafeURL := validation.IsSafeURL
	// Mock IsSafeURL to simulate real behavior (since TestMain disables it)
	validation.IsSafeURL = func(urlStr string) error {
		if strings.Contains(urlStr, "127.0.0.1") || strings.Contains(urlStr, "localhost") {
			return fmt.Errorf("loopback address is not allowed")
		}
		return nil
	}
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	ctx := context.Background()

	// Define a tool that simulates curl
	cmdService := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"), // Use echo for safety, but imagine it's curl
		// Args are part of CallDefinition, not Service in the builder used in other tests?
		// Wait, command_coverage_test.go uses callDef for Args.
		// But UpstreamServiceConfig has CommandLineUpstreamService which has Args?
        // Let's check command_coverage_test.go again.
        // It sets Args in callDef.
        // But UpstreamServiceConfig usually defines the base command.
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

	pbTool := pb.Tool_builder{
		Name: proto.String("fetch_url"),
	}.Build()

	targetTool := tool.NewLocalCommandTool(pbTool, cmdService, callDef, nil, "call1")

	// 1. Private IP (Loopback)
	// Currently, this passes because there is no SSRF check.
	reqPrivate := &tool.ExecutionRequest{
		ToolName:   "fetch_url",
		ToolInputs: []byte(`{"url": "http://127.0.0.1/metadata"}`),
	}

	_, err := targetTool.Execute(ctx, reqPrivate)

	// Expectation: Failure (Secure) due to SSRF protection
	assert.Error(t, err, "Should fail due to SSRF protection")
	if err != nil {
		// Can fail via IsSafeURL check ("unsafe url argument") or checkForSSRF ("potential SSRF detected")
		isUnsafe := strings.Contains(err.Error(), "unsafe url argument")
		isSSRF := strings.Contains(err.Error(), "potential SSRF detected")
		assert.True(t, isUnsafe || isSSRF, "Error should indicate security violation: %v", err)
	}

	// 2. Public IP
	reqPublic := &tool.ExecutionRequest{
		ToolName:   "fetch_url",
		ToolInputs: []byte(`{"url": "http://example.com"}`),
	}
	resPublic, errPublic := targetTool.Execute(ctx, reqPublic)
	assert.NoError(t, errPublic)
	_ = resPublic

	// 3. Raw IP (127.0.0.1)
	reqIP := &tool.ExecutionRequest{
		ToolName:   "fetch_url",
		ToolInputs: []byte(`{"url": "127.0.0.1"}`),
	}
	_, errIP := targetTool.Execute(ctx, reqIP)
	assert.Error(t, errIP)
	if errIP != nil {
		assert.Contains(t, errIP.Error(), "potential SSRF detected")
	}

	// 4. Localhost string
	reqLocalhost := &tool.ExecutionRequest{
		ToolName:   "fetch_url",
		ToolInputs: []byte(`{"url": "localhost"}`),
	}
	_, errLocalhost := targetTool.Execute(ctx, reqLocalhost)
	assert.Error(t, errLocalhost)
	if errLocalhost != nil {
		assert.Contains(t, errLocalhost.Error(), "potential SSRF detected")
	}
}
