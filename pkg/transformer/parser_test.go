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

func TestTextParser_ParseJSON_ErrorCases(t *testing.T) {
	parser := NewTextParser()
	jsonInput := []byte(`{"person": {"name": "test", "age": 123}}`)

	t.Run("invalid_jsonpath", func(t *testing.T) {
		config := map[string]string{"name": `{.person.name[`}
		_, err := parser.Parse("json", jsonInput, config)
		assert.Error(t, err)
	})

	t.Run("jsonpath_not_found", func(t *testing.T) {
		config := map[string]string{"name": `{.person.nonexistent}`}
		result, err := parser.Parse("json", jsonInput, config)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestTextParser_GeneralEdgeCases(t *testing.T) {
	parser := NewTextParser()

	t.Run("unsupported_type", func(t *testing.T) {
		_, err := parser.Parse("csv", []byte{}, nil)
		assert.Error(t, err)
	})

	t.Run("empty_input_and_config", func(t *testing.T) {
		result, err := parser.Parse("json", []byte(`{}`), map[string]string{})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestTextParser_ParseXML_ErrorCases(t *testing.T) {
	parser := NewTextParser()
	xmlInput := []byte(`<root><name>test</name><value>123</value></root>`)

	t.Run("invalid_xpath", func(t *testing.T) {
		config := map[string]string{"name": `//name[`}
		_, err := parser.Parse("xml", xmlInput, config)
		assert.Error(t, err)
	})

	t.Run("invalid_xml", func(t *testing.T) {
		invalidXMLInput := []byte(`<root><name>test</name>`)
		config := map[string]string{"name": `//name`}
		_, err := parser.Parse("xml", invalidXMLInput, config)
		assert.Error(t, err)
	})

	t.Run("xpath_not_found", func(t *testing.T) {
		config := map[string]string{"name": `//nonexistent`}
		result, err := parser.Parse("xml", xmlInput, config)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestTextParser_ParseJSON_EdgeCases(t *testing.T) {
	parser := NewTextParser()

	t.Run("empty_config", func(t *testing.T) {
		jsonInput := []byte(`{"person": {"name": "test"}}`)
		config := map[string]string{}
		result, err := parser.Parse("json", jsonInput, config)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty_input", func(t *testing.T) {
		jsonInput := []byte(`{}`)
		config := map[string]string{"name": `{.person.name}`}
		result, err := parser.Parse("json", jsonInput, config)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestTextParser_ParseJSON_Complex(t *testing.T) {
	parser := NewTextParser()
	jsonInput := []byte(`{
		"person": {
			"name": "test",
			"age": 123,
			"contacts": [
				{"type": "email", "value": "test@example.com"},
				{"type": "phone", "value": "123-456-7890"}
			]
		}
	}`)
	config := map[string]string{
		"name":         `{.person.name}`,
		"age":          `{.person.age}`,
		"email":        `{.person.contacts[?(@.type=="email")].value}`,
		"firstContact": `{.person.contacts[0].value}`,
	}
	result, err := parser.Parse("json", jsonInput, config)
	require.NoError(t, err)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(123), result["age"])
	assert.Equal(t, "test@example.com", result["email"])
	assert.Equal(t, "test@example.com", result["firstContact"])
}

func TestTextParser_ParseXML_WithNamespaces(t *testing.T) {
	parser := NewTextParser()
	xmlInput := []byte(`
		<root xmlns:h="http://www.w3.org/TR/html4/">
			<h:table>
				<h:tr>
					<h:td>Apples</h:td>
					<h:td>Bananas</h:td>
				</h:tr>
			</h:table>
		</root>
	`)
	config := map[string]string{
		"cell1": `//h:td[1]`,
		"cell2": `//h:td[2]`,
	}
	result, err := parser.Parse("xml", xmlInput, config)
	require.NoError(t, err)
	assert.Equal(t, "Apples", result["cell1"])
	assert.Equal(t, "Bananas", result["cell2"])
}

func TestTextParser_ParseText_Complex(t *testing.T) {
	parser := NewTextParser()
	textInput := []byte(`Event: "user_login", Status: "success", User-ID: "user-123@example.com"`)
	config := map[string]string{
		"event":  `Event: "([^"]+)"`,
		"status": `Status: "(\w+)"`,
		"email":  `User-ID: "([^"]+)"`,
	}
	result, err := parser.Parse("text", textInput, config)
	require.NoError(t, err)
	assert.Equal(t, "user_login", result["event"])
	assert.Equal(t, "success", result["status"])
	assert.Equal(t, "user-123@example.com", result["email"])
}

func TestTextParser_ParseText_ErrorCases(t *testing.T) {
	parser := NewTextParser()
	textInput := []byte(`User ID: 12345, Name: John Doe`)

	t.Run("invalid_regex", func(t *testing.T) {
		config := map[string]string{"userId": `User ID: (\d+[`}
		_, err := parser.Parse("text", textInput, config)
		assert.Error(t, err)
	})

	t.Run("regex_not_found", func(t *testing.T) {
		config := map[string]string{"userId": `User ID: (\s+)`}
		result, err := parser.Parse("text", textInput, config)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}
