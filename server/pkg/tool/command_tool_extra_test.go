// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestLocalCommandTool_PolicyBlock(t *testing.T) {
	// Setup tool with denying policy
	deny := configv1.CallPolicy_DENY
	policies := []*configv1.CallPolicy{
		{
			DefaultAction: &deny,
		},
	}

	lct := NewLocalCommandTool(
		&v1.Tool{Name: proto.String("test")},
		&configv1.CommandLineUpstreamService{Command: proto.String("echo")},
		&configv1.CommandLineCallDefinition{},
		policies,
		"call1",
	)

	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	}

	_, err := lct.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution blocked by policy")
}

func TestLocalCommandTool_ShellInjection(t *testing.T) {
	lct := NewLocalCommandTool(
		&v1.Tool{Name: proto.String("test")},
		&configv1.CommandLineUpstreamService{Command: proto.String("bash")}, // Shell command
		&configv1.CommandLineCallDefinition{
			Args: []string{"-c", "{{script}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("script")},
				},
			},
		},
		nil,
		"call1",
	)

	// Attempt injection
	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"script": "echo hello; rm -rf /"}`),
	}

	_, err := lct.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
}

func TestLocalCommandTool_ShellInjection_DoubleQuote(t *testing.T) {
	lct := NewLocalCommandTool(
		&v1.Tool{Name: proto.String("test")},
		&configv1.CommandLineUpstreamService{Command: proto.String("bash")}, // Shell command
		&configv1.CommandLineCallDefinition{
			Args: []string{"-c", "\"{{script}}\""},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("script")},
				},
			},
		},
		nil,
		"call1",
	)

	// Attempt injection with backtick
	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"script": "` + "`whoami`" + `"}`),
	}

	_, err := lct.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
}
