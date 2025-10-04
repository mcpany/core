/*
 * Copyright 2025 Author(s) of MCP-XY
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

func TestTextParser_ParseJSON(t *testing.T) {
	parser := NewTextParser()
	jsonInput := []byte(`{"person": {"name": "test", "age": 123}}`)
	config := map[string]string{
		"name": `{.person.name}`,
		"age":  `{.person.age}`,
	}
	result, err := parser.Parse("json", jsonInput, config)
	require.NoError(t, err)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(123), result["age"]) // jsonpath returns float64 for numbers
}

func TestTextParser_ParseInvalidJSON(t *testing.T) {
	parser := NewTextParser()
	jsonInput := []byte(`{"name": "test"`)
	_, err := parser.Parse("json", jsonInput, nil)
	require.Error(t, err)
}

func TestTextParser_ParseXML(t *testing.T) {
	parser := NewTextParser()
	xmlInput := []byte(`<root><name>test</name><value>123</value></root>`)
	config := map[string]string{
		"name":  `//name`,
		"value": `//value`,
	}
	result, err := parser.Parse("xml", xmlInput, config)
	require.NoError(t, err)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "123", result["value"])
}

func TestTextParser_ParseText(t *testing.T) {
	parser := NewTextParser()
	textInput := []byte(`User ID: 12345, Name: John Doe`)
	config := map[string]string{
		"userId": `User ID: (\d+)`,
		"name":   `Name: ([\w\s]+)`,
	}
	result, err := parser.Parse("text", textInput, config)
	require.NoError(t, err)
	assert.Equal(t, "12345", result["userId"])
	assert.Equal(t, "John Doe", result["name"])
}

func TestTextParser_UnsupportedType(t *testing.T) {
	parser := NewTextParser()
	_, err := parser.Parse("yaml", []byte{}, nil)
	require.Error(t, err)
}
