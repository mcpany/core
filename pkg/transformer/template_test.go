/*
 * Copyright 2025 Author(s) of MCPXY
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

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextTemplate_Render(t *testing.T) {
	templateString := "Hello, {{.name}}! You are {{.age}} years old."
	tpl, err := NewTextTemplate(templateString)
	require.NoError(t, err)

	params := map[string]any{
		"name": "World",
		"age":  99,
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World! You are 99 years old.", rendered)
}

func TestTextTemplate_InvalidTemplate(t *testing.T) {
	templateString := "Hello, {{.name!"
	_, err := NewTextTemplate(templateString)
	require.Error(t, err)
}

func TestTextTemplate_MissingParameter(t *testing.T) {
	templateString := "Hello, {{.name}}!"
	tpl, err := NewTextTemplate(templateString)
	require.NoError(t, err)

	params := map[string]any{}
	rendered, err := tpl.Render(params)
	require.NoError(t, err)
	assert.Equal(t, "Hello, <no value>!", rendered)
}
