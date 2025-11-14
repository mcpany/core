// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e
// +build e2e

package integration

import (
	"context"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestToolAndResourceWithPrompts(t *testing.T) {
	t.Parallel()

	// Start the MCP Any server
	serverInfo := StartMCPANYServer(t, "prompts-test")
	defer serverInfo.CleanupFunc()

	// Define a tool with a prompt
	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("test-tool-with-prompt"),
		Description: proto.String("A test tool with a prompt"),
		Prompts: []*configv1.PromptDefinition{
			configv1.PromptDefinition_builder{
				Uri:  proto.String("mcp://prompts/test-prompt"),
				Text: proto.String("This is a test prompt for a tool."),
			}.Build(),
		},
		CallId: proto.String("test-call"),
	}.Build()

	// Define a resource with a prompt
	resDef := configv1.ResourceDefinition_builder{
		Uri:         proto.String("mcp://resources/test-resource-with-prompt"),
		Description: proto.String("A test resource with a prompt"),
		Prompts: []*configv1.PromptDefinition{
			configv1.PromptDefinition_builder{
				Uri:  proto.String("mcp://prompts/test-prompt-resource"),
				Text: proto.String("This is a test prompt for a resource."),
			}.Build(),
		},
		Static: configv1.StaticResource_builder{
			TextContent: proto.String("This is a test resource."),
		}.Build(),
	}.Build()

	// Define a dummy call definition
	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	callDef := configv1.HttpCallDefinition_builder{
		Id:     proto.String("test-call"),
		Method: &method,
	}.Build()

	// Register a service with the tool and resource
	upstreamServiceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-with-prompts"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address:   proto.String("http://localhost:8080"),
			Tools:     []*configv1.ToolDefinition{toolDef},
			Resources: []*configv1.ResourceDefinition{resDef},
			Calls:     map[string]*configv1.HttpCallDefinition{"test-call": callDef},
		}.Build(),
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: upstreamServiceConfig,
	}.Build()

	RegisterServiceViaAPI(t, serverInfo.RegistrationClient, req)

	// List the prompts and verify that they are registered
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.Eventually(t, func() bool {
		prompts, err := serverInfo.ListPrompts(ctx)
		if err != nil {
			t.Logf("Error listing prompts: %v", err)
			return false
		}
		if len(prompts.Prompts) != 2 {
			t.Logf("Expected 2 prompts, got %d", len(prompts.Prompts))
			return false
		}
		return true
	}, 5*time.Second, 250*time.Millisecond, "Prompts were not registered in time")

	prompts, err := serverInfo.ListPrompts(ctx)
	require.NoError(t, err)

	assert.Len(t, prompts.Prompts, 2)

	var toolPromptFound, resourcePromptFound bool
	for _, p := range prompts.Prompts {
		if p.Name == "test-prompt" {
			toolPromptFound = true
			assert.Equal(t, "This is a test prompt for a tool.", p.Description)
		}
		if p.Name == "test-prompt-resource" {
			resourcePromptFound = true
			assert.Equal(t, "This is a test prompt for a resource.", p.Description)
		}
	}
	assert.True(t, toolPromptFound, "Tool prompt not found")
	assert.True(t, resourcePromptFound, "Resource prompt not found")
}
