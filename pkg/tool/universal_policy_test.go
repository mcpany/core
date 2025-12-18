// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUniversalPolicyEnforcement(t *testing.T) {
	// Create a policy that blocks everything with "blocked" in arguments
	policy := &configv1.CallPolicy{
		DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
		Rules: []*configv1.CallPolicyRule{
			{
				ArgumentRegex: proto.String(".*blocked.*"),
				Action:        configv1.CallPolicy_DENY.Enum(),
			},
		},
	}
	policies := []*configv1.CallPolicy{policy}

	t.Run("GRPCTool", func(t *testing.T) {
		toolProto := &v1.Tool{Name: proto.String("grpc-tool")}
		pm := pool.NewManager()
		mockMethodDesc := new(MockMethodDescriptor)
		mockMsgDesc := new(MockMessageDescriptor)
		mockMethodDesc.On("Input").Return(mockMsgDesc)

		// Create tool with policy
		grpcTool := NewGRPCTool(toolProto, pm, "service", mockMethodDesc, nil, policies, "")

		// Allowed call
		_, err := grpcTool.Execute(context.Background(), &ExecutionRequest{
			ToolName:   "grpc-tool",
			ToolInputs: json.RawMessage(`{"arg":"safe"}`),
		})
		// It will fail later because of missing connection pool, but NOT "blocked by policy"
		if err != nil {
			assert.NotContains(t, err.Error(), "blocked by policy")
		}

		// Blocked call
		_, err = grpcTool.Execute(context.Background(), &ExecutionRequest{
			ToolName:   "grpc-tool",
			ToolInputs: json.RawMessage(`{"arg":"blocked"}`),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by policy")
	})

	t.Run("OpenAPITool", func(t *testing.T) {
		toolProto := &v1.Tool{Name: proto.String("openapi-tool")}
		// client can be nil if policy checks fail first?
		// NewOpenAPITool doesn't use client in constructor.

		openapiTool := NewOpenAPITool(toolProto, nil, nil, "GET", "http://example.com", nil, nil, policies, "")

		// Blocked call
		_, err := openapiTool.Execute(context.Background(), &ExecutionRequest{
			ToolName:   "openapi-tool",
			ToolInputs: json.RawMessage(`{"arg":"blocked"}`),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by policy")
	})

	t.Run("CommandTool", func(t *testing.T) {
		toolProto := &v1.Tool{Name: proto.String("cmd-tool")}
		cmdService := &configv1.CommandLineUpstreamService{
			Command: proto.String("echo"),
		}

		cmdTool := NewCommandTool(toolProto, cmdService, nil, policies, "")

		// Blocked call
		_, err := cmdTool.Execute(context.Background(), &ExecutionRequest{
			ToolName:   "cmd-tool",
			ToolInputs: json.RawMessage(`{"arg":"blocked"}`),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by policy")
	})

	t.Run("LocalCommandTool", func(t *testing.T) {
		toolProto := &v1.Tool{Name: proto.String("local-cmd-tool")}
		cmdService := &configv1.CommandLineUpstreamService{
			Command: proto.String("echo"),
		}

		localCmdTool := NewLocalCommandTool(toolProto, cmdService, nil, policies, "")

		// Blocked call
		_, err := localCmdTool.Execute(context.Background(), &ExecutionRequest{
			ToolName:   "local-cmd-tool",
			ToolInputs: json.RawMessage(`{"arg":"blocked"}`),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by policy")
	})

	t.Run("MCPTool", func(t *testing.T) {
		toolProto := &v1.Tool{Name: proto.String("mcp-tool")}

		mcpTool := NewMCPTool(toolProto, nil, &configv1.MCPCallDefinition{}, policies, "")

		// Blocked call
		_, err := mcpTool.Execute(context.Background(), &ExecutionRequest{
			ToolName:   "mcp-tool",
			ToolInputs: json.RawMessage(`{"arg":"blocked"}`),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by policy")
	})
}
