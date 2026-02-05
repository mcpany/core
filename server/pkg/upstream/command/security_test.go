// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestArgInjection(t *testing.T) {
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()
	u := NewUpstream()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-vuln-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("/bin/echo"),
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"echo-call": configv1.CommandLineCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("echo-safe"),
					CallId: proto.String("echo-call"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(
		context.Background(),
		serviceConfig,
		tm,
		prm,
		rm,
		false,
	)
	require.NoError(t, err)

	cmdTool := tm.ListTools()[0]

	// Check if args is in schema (it shouldn't be for a safe tool)
	inputSchema := cmdTool.Tool().GetInputSchema()
	fields := inputSchema.Fields["properties"].GetStructValue().GetFields()
	_, hasArgs := fields["args"]
	assert.False(t, hasArgs, "args parameter should not be automatically added")

	// Try to inject args
	inputData := map[string]interface{}{"args": []string{"injected"}}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// This should now FAIL
	_, err = cmdTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'args' parameter is not allowed")
}
