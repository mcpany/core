package mcpserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpToolManager(t *testing.T) {
	m := &NoOpToolManager{}

	err := m.AddTool(nil)
	assert.NoError(t, err)

	tool, ok := m.GetTool("test")
	assert.Nil(t, tool)
	assert.False(t, ok)

	tools := m.ListTools()
	assert.Nil(t, tools)

	mcpTools := m.ListMCPTools()
	assert.Nil(t, mcpTools)

	m.ClearToolsForService("test")

	res, err := m.ExecuteTool(context.Background(), nil)
	assert.Nil(t, res)
	assert.NoError(t, err)

	m.SetMCPServer(nil)
	m.AddMiddleware(nil)
	m.AddServiceInfo("test", nil)

	info, ok := m.GetServiceInfo("test")
	assert.Nil(t, info)
	assert.False(t, ok)

	services := m.ListServices()
	assert.Nil(t, services)

	m.SetProfiles(nil, nil)

	allowed := m.IsServiceAllowed("test", "test")
	assert.True(t, allowed)

	matches := m.ToolMatchesProfile(nil, "test")
	assert.True(t, matches)

	ids, ok := m.GetAllowedServiceIDs("test")
	assert.Nil(t, ids)
	assert.False(t, ok)

	count := m.GetToolCountForService("test")
	assert.Equal(t, 0, count)
}

func TestNoOpPromptManager(t *testing.T) {
	m := &NoOpPromptManager{}

	m.AddPrompt(nil)
	m.UpdatePrompt(nil)

	prompt, ok := m.GetPrompt("test")
	assert.Nil(t, prompt)
	assert.False(t, ok)

	prompts := m.ListPrompts()
	assert.Nil(t, prompts)

	m.ClearPromptsForService("test")
	m.SetMCPServer(nil)
}

func TestNoOpResourceManager(t *testing.T) {
	m := &NoOpResourceManager{}

	res, ok := m.GetResource("test")
	assert.Nil(t, res)
	assert.False(t, ok)

	m.AddResource(nil)
	m.RemoveResource("test")

	resources := m.ListResources()
	assert.Nil(t, resources)

	m.OnListChanged(nil)
	m.ClearResourcesForService("test")
}
