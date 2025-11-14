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

package prompt

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguredPrompt(t *testing.T) {
	def := &configv1.PromptDefinition{}
	def.SetName("test-prompt")
	def.SetTitle("Test Prompt")
	def.SetDescription("A prompt for testing.")
	pa := &configv1.PromptArgument{}
	pa.SetName("arg1")
	pa.SetDescription("Argument 1")
	pa.SetRequired(true)
	def.SetArguments([]*configv1.PromptArgument{pa})
	pm := &configv1.PromptMessage{}
	pm.SetRole(configv1.PromptMessage_USER)
	tc := &configv1.TextContent{}
	tc.SetText("Hello, {{.arg1}}!")
	pm.SetText(tc)
	def.SetMessages([]*configv1.PromptMessage{pm})

	p, err := NewConfiguredPrompt(def, "test-service")
	require.NoError(t, err)

	assert.Equal(t, "test-prompt", p.Prompt().Name)
	assert.Equal(t, "Test Prompt", p.Prompt().Title)
	assert.Equal(t, "A prompt for testing.", p.Prompt().Description)
	assert.Len(t, p.Prompt().Arguments, 1)
	assert.Equal(t, "arg1", p.Prompt().Arguments[0].Name)
	assert.Equal(t, "test-service", p.Service())

	args := json.RawMessage(`{"arg1": "world"}`)
	result, err := p.Get(context.Background(), args)
	require.NoError(t, err)
	require.Len(t, result.Messages, 1)
	assert.Equal(t, mcp.Role("user"), result.Messages[0].Role)
	textContent, ok := result.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, world!", textContent.Text)
}
