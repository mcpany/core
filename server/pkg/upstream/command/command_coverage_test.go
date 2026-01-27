// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestUpstream_Register_DisabledItems(t *testing.T) {
	u := NewUpstream()
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("disabled-items-service"),
	}
	cmdService := &configv1.CommandLineUpstreamService{
		Tools: []*configv1.ToolDefinition{
			{Name: proto.String("tool1"), Disable: proto.Bool(true)},
		},
		Prompts: []*configv1.PromptDefinition{
			{Name: proto.String("prompt1"), Disable: proto.Bool(true)},
		},
		Resources: []*configv1.ResourceDefinition{
			{Name: proto.String("resource1"), Disable: proto.Bool(true)},
		},
	}
	config.SetCommandLineService(cmdService)

	_, tools, _, err := u.Register(context.Background(), config, tm, prm, rm, false)
	require.NoError(t, err)
	assert.Empty(t, tools)
	assert.Empty(t, tm.ListTools())
	assert.Empty(t, prm.ListPrompts())
	assert.Empty(t, rm.ListResources())
}

func TestUpstream_Register_DynamicResourceErrors(t *testing.T) {
	u := NewUpstream()
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("dynamic-errors"),
	}

	cmdService := &configv1.CommandLineUpstreamService{
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"c1": {Id: proto.String("c1")},
		},
	}

	resDef := &configv1.ResourceDefinition{
		Name: proto.String("res1"),
	}
	dynRes := &configv1.DynamicResource{}
	dynRes.SetCommandLineCall(&configv1.CommandLineCallDefinition{Id: proto.String("c1")})
	resDef.SetDynamic(dynRes)

	cmdService.SetResources([]*configv1.ResourceDefinition{resDef})
	config.SetCommandLineService(cmdService)

	// Register should succeed but resource not added because toolName not found for c1
	_, _, _, err := u.Register(context.Background(), config, tm, prm, rm, false)
	require.NoError(t, err)
	assert.Empty(t, rm.ListResources())
}

func TestUpstream_Register_DisabledPrompt(_ *testing.T) {
	// Covered in DisabledItems
}
