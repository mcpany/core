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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPromptManager(t *testing.T) {
	pm := NewPromptManager()
	assert.NotNil(t, pm)
	assert.NotNil(t, pm.prompts)
}

func TestPromptManager_GetListClear(t *testing.T) {
	pm := NewPromptManager()

	// Add some prompts
	pm.prompts["prompt1"] = &Prompt{Name: "prompt1", ServiceID: "service1"}
	pm.prompts["prompt2"] = &Prompt{Name: "prompt2", ServiceID: "service2"}

	// Get
	p, ok := pm.GetPrompt("prompt1")
	assert.True(t, ok)
	assert.Equal(t, "prompt1", p.Name)

	p, ok = pm.GetPrompt("prompt2")
	assert.True(t, ok)
	assert.Equal(t, "prompt2", p.Name)

	_, ok = pm.GetPrompt("non-existent")
	assert.False(t, ok)

	// List
	prompts := pm.ListPrompts()
	assert.Len(t, prompts, 2)

	// Clear
	pm.ClearPromptsForService("service1")
	assert.Len(t, pm.ListPrompts(), 1)
	pm.ClearPromptsForService("service2")
	assert.Len(t, pm.ListPrompts(), 0)
}
