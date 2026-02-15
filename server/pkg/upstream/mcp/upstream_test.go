package mcp

import (
	"context"
	"net/http"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

type mockToolManager struct {
	tool.ManagerInterface
	tools map[string]tool.Tool
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	m.tools[t.Tool().GetName()] = t
	return nil
}

func (m *mockToolManager) GetTool(toolName string) (tool.Tool, bool) {
	t, ok := m.tools[toolName]
	return t, ok
}

func (m *mockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

type mockPromptManager struct {
	prompt.ManagerInterface
	prompts map[string]prompt.Prompt
}

func newMockPromptManager() *mockPromptManager {
	return &mockPromptManager{
		prompts: make(map[string]prompt.Prompt),
	}
}

func (m *mockPromptManager) AddPrompt(p prompt.Prompt) {
	m.prompts[p.Prompt().Name] = p
}

func (m *mockPromptManager) GetPrompt(name string) (prompt.Prompt, bool) {
	p, ok := m.prompts[name]
	return p, ok
}

type mockResourceManager struct {
	resource.ManagerInterface
	resources map[string]resource.Resource
}

func newMockResourceManager() *mockResourceManager {
	return &mockResourceManager{
		resources: make(map[string]resource.Resource),
	}
}

func (m *mockResourceManager) AddResource(r resource.Resource) {
	m.resources[r.Resource().URI] = r
}

func (m *mockResourceManager) GetResource(uri string) (resource.Resource, bool) {
	r, ok := m.resources[uri]
	return r, ok
}

func (m *mockResourceManager) OnListChanged(_ func()) {}

func (m *mockResourceManager) Subscribe(_ context.Context, _ string) error {
	return nil
}

type mockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

func (m *mockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}
