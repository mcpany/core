// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_SSRF_SCP_Syntax(t *testing.T) {
	// Unset the bypass environment variable for this test
	originalEnv := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")
	defer os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", originalEnv)

	// This test verifies that we detect and block SCP-style URLs that resolve to private IPs.
	// e.g. git clone git@private.internal:repo.git

	// Mock DNS resolution to return a private IP for our target host
	originalLookup := validation.LookupIPFunc
	defer func() { validation.LookupIPFunc = originalLookup }()

	validation.LookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
		if host == "private.internal" {
			return []net.IP{net.ParseIP("192.168.1.1")}, nil
		}
		return []net.IP{}, fmt.Errorf("unknown host")
	}

	toolDef := configv1.ToolDefinition_builder{
		Name: proto.String("git-clone"),
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("git-clone"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
		Tools:   []*configv1.ToolDefinition{toolDef},
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}", "/tmp/repo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	policies := []*configv1.CallPolicy{} // No specific policy

	localTool := NewLocalCommandTool(toolProto, service, callDef, policies, "call-id")

	// Payload: SCP-style URL pointing to private IP
	payload := "git@private.internal:repo.git"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		ToolInputs: []byte(fmt.Sprintf(`{"url": "%s"}`, payload)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := localTool.Execute(ctx, req)

	// We expect a validation error now.
	assert.Error(t, err, "Expected execution to fail with validation error")
	if err != nil {
		// The error message should indicate unsafe argument or private network
		// My code returns "unsafe git argument: ... private network address is not allowed"
		assert.Contains(t, err.Error(), "unsafe git argument", "Error should mention unsafe git argument")
		assert.Contains(t, err.Error(), "private network", "Error should mention private network")
	}
}
