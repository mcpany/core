/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"net/http"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
)

type mockToolManager struct {
	tool.ToolManagerInterface
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

func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}

type mockPromptManager struct {
	prompt.PromptManagerInterface
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
	resource.ResourceManagerInterface
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

func (m *mockResourceManager) OnListChanged(f func()) {}

func (m *mockResourceManager) Subscribe(ctx context.Context, uri string) error {
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
	return &http.Response{StatusCode: http.StatusOK}, nil
}
