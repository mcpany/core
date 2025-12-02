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

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextParser_ParseAndTransform_JSON(t *testing.T) {
	parser := NewTextParser()
	jsonInput := []byte(`{"person": {"name": "test", "age": 123}}`)
	config := map[string]string{
		"name": `{.person.name}`,
		"age":  `{.person.age}`,
	}
	template := `{"name": "{{.name}}", "age": {{.age}}}`

	parsed, err := parser.Parse("json", jsonInput, config)
	require.NoError(t, err)

	result, err := parser.Transform(template, parsed)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name": "test", "age": 123}`, string(result))
}

func TestTextParser_ParseAndTransform_XML(t *testing.T) {
	parser := NewTextParser()
	xmlInput := []byte(`<root><name>test</name><value>123</value></root>`)
	config := map[string]string{
		"name":  `//name`,
		"value": `//value`,
	}
	template := `{"name": "{{.name}}", "value": "{{.value}}"}`

	parsed, err := parser.Parse("xml", xmlInput, config)
	require.NoError(t, err)

	result, err := parser.Transform(template, parsed)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name": "test", "value": "123"}`, string(result))
}

func TestTextParser_ParseAndTransform_Text(t *testing.T) {
	parser := NewTextParser()
	textInput := []byte(`User ID: 12345, Name: John Doe`)
	config := map[string]string{
		"userId": `User ID: (\d+)`,
		"name":   `Name: ([\w\s]+)`,
	}
	template := `{"userId": "{{.userId}}", "name": "{{.name}}"}`

	parsed, err := parser.Parse("text", textInput, config)
	require.NoError(t, err)

	result, err := parser.Transform(template, parsed)
	require.NoError(t, err)
	assert.JSONEq(t, `{"userId": "12345", "name": "John Doe"}`, string(result))
}
