package mcpserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpToolManager(t *testing.T) {
	tm := &NoOpToolManager{}

	assert.NoError(t, tm.AddTool(nil))

	tool, ok := tm.GetTool("any")
	assert.Nil(t, tool)
	assert.False(t, ok)

	assert.Nil(t, tm.ListTools())
	assert.Nil(t, tm.ListMCPTools())

	tm.ClearToolsForService("any")

	res, err := tm.ExecuteTool(context.Background(), nil)
	assert.Nil(t, res)
	assert.NoError(t, err)

	tm.SetMCPServer(nil)
	tm.AddMiddleware(nil)
	tm.AddServiceInfo("any", nil)

	info, ok := tm.GetServiceInfo("any")
	assert.Nil(t, info)
	assert.False(t, ok)

	assert.Nil(t, tm.ListServices())

	tm.SetProfiles(nil, nil)

	assert.True(t, tm.IsServiceAllowed("any", "any"))
	assert.True(t, tm.ToolMatchesProfile(nil, "any"))

	allowed, ok := tm.GetAllowedServiceIDs("any")
	assert.Nil(t, allowed)
	assert.False(t, ok)

	assert.Equal(t, 0, tm.GetToolCountForService("any"))
}

func TestNoOpPromptManager(t *testing.T) {
	pm := &NoOpPromptManager{}

	pm.AddPrompt(nil)
	pm.UpdatePrompt(nil)

	prompt, ok := pm.GetPrompt("any")
	assert.Nil(t, prompt)
	assert.False(t, ok)

	assert.Nil(t, pm.ListPrompts())

	pm.ClearPromptsForService("any")
	pm.SetMCPServer(nil)
}

func TestNoOpResourceManager(t *testing.T) {
	rm := &NoOpResourceManager{}

	res, ok := rm.GetResource("any")
	assert.Nil(t, res)
	assert.False(t, ok)

	rm.AddResource(nil)
	rm.RemoveResource("any")

	assert.Nil(t, rm.ListResources())

	rm.OnListChanged(func() {})
	rm.ClearResourcesForService("any")
}
