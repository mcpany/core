// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func newToolDef(name string, disabled bool) *configv1.ToolDefinition {
	t := &configv1.ToolDefinition{}
	t.SetName(name)
	t.SetDisable(disabled)
	return t
}

func newPromptDef(name string, disabled bool) *configv1.PromptDefinition {
	p := &configv1.PromptDefinition{}
	p.SetName(name)
	p.SetDisable(disabled)
	return p
}

func newResourceDef(name string, disabled bool) *configv1.ResourceDefinition {
	r := &configv1.ResourceDefinition{}
	r.SetName(name)
	r.SetDisable(disabled)
	return r
}

func newCmdCall(id string) *configv1.CommandLineCallDefinition {
	c := &configv1.CommandLineCallDefinition{}
	c.SetId(id)
	return c
}

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
			newToolDef("tool1", true),
		},
		Prompts: []*configv1.PromptDefinition{
			newPromptDef("prompt1", true),
		},
		Resources: []*configv1.ResourceDefinition{
			newResourceDef("resource1", true),
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
			"c1": newCmdCall("c1"),
		},
	}

	resDef := &configv1.ResourceDefinition{}
	resDef.SetName("res1")

	dynRes := &configv1.DynamicResource{}
	dynRes.SetCommandLineCall(newCmdCall("c1"))
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
