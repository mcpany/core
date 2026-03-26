// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextParser_ParseJSON(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	jsonInput := []byte(`{"person": {"name": "test", "age": 123}}`)
	config := map[string]string{
		"name": `{.person.name}`,
		"age":  `{.person.age}`,
	}
	result, err := parser.Parse("json", jsonInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)
	assert.Equal(t, "test", resMap["name"])
	assert.Equal(t, float64(123), resMap["age"]) // jsonpath returns float64 for numbers
}

func TestTextParser_ParseInvalidJSON(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	jsonInput := []byte(`{"name": "test"`)
	_, err := parser.Parse("json", jsonInput, nil, "")
	require.Error(t, err)
}

func TestTextParser_ParseXML(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	xmlInput := []byte(`<root><name>test</name><value>123</value></root>`)
	config := map[string]string{
		"name":  `//name`,
		"value": `//value`,
	}
	result, err := parser.Parse("xml", xmlInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)
	assert.Equal(t, "test", resMap["name"])
	assert.Equal(t, "123", resMap["value"])
}

func TestTextParser_ParseText(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	textInput := []byte(`User ID: 12345, Name: John Doe`)
	config := map[string]string{
		"userId": `User ID: (\d+)`,
		"name":   `Name: ([\w\s]+)`,
	}
	result, err := parser.Parse("text", textInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)
	assert.Equal(t, "12345", resMap["userId"])
	assert.Equal(t, "John Doe", resMap["name"])
}

func TestTextParser_UnsupportedType(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	_, err := parser.Parse("yaml", []byte{}, nil, "")
	require.Error(t, err)
}

func TestTextParser_ParseJSON_ErrorCases(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	jsonInput := []byte(`{"person": {"name": "test", "age": 123}}`)

	t.Run("invalid_jsonpath", func(t *testing.T) {
		config := map[string]string{"name": `{.person.name[`}
		_, err := parser.Parse("json", jsonInput, config, "")
		assert.Error(t, err)
	})

	t.Run("jsonpath_not_found", func(t *testing.T) {
		config := map[string]string{"name": `{.person.nonexistent}`}
		result, err := parser.Parse("json", jsonInput, config, "")
		assert.NoError(t, err)
		resMap := result.(map[string]any)
		assert.Empty(t, resMap)
	})
}

func TestTextParser_ParseXML_ErrorCases(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	xmlInput := []byte(`<root><name>test</name><value>123</value></root>`)

	t.Run("invalid_xpath", func(t *testing.T) {
		config := map[string]string{"name": `//name[`}
		_, err := parser.Parse("xml", xmlInput, config, "")
		assert.Error(t, err)
	})
}

func TestTextParser_Transform(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	data := map[string]any{
		"name": "test",
		"age":  123,
	}
	template := `{"name": "{{.name}}", "age": {{.age}}}`
	result, err := parser.Transform(template, data)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name": "test", "age": 123}`, string(result))
}

func TestTextParser_ParseJSON_Complex(t *testing.T) {
	t.Parallel()
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
	result, err := parser.Parse("json", jsonInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)
	assert.Equal(t, "test", resMap["name"])
	assert.Equal(t, float64(123), resMap["age"])
	assert.Equal(t, "test@example.com", resMap["email"])
	assert.Equal(t, "test@example.com", resMap["firstContact"])
}

func TestTextParser_ParseXML_WithNamespaces(t *testing.T) {
	t.Parallel()
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
	result, err := parser.Parse("xml", xmlInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)
	assert.Equal(t, "Apples", resMap["cell1"])
	assert.Equal(t, "Bananas", resMap["cell2"])
}

func TestTextParser_ParseText_Complex(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	textInput := []byte(`Event: "user_login", Status: "success", User-ID: "user-123@example.com"`)
	config := map[string]string{
		"event":  `Event: "([^"]+)"`,
		"status": `Status: "(\w+)"`,
		"email":  `User-ID: "([^"]+)"`,
	}
	result, err := parser.Parse("text", textInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)
	assert.Equal(t, "user_login", resMap["event"])
	assert.Equal(t, "success", resMap["status"])
	assert.Equal(t, "user-123@example.com", resMap["email"])
}

func TestTextParser_ParseText_ErrorCases(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	textInput := []byte(`User ID: 12345, Name: John Doe`)

	t.Run("invalid_regex", func(t *testing.T) {
		config := map[string]string{"userId": `User ID: (\d+[`}
		_, err := parser.Parse("text", textInput, config, "")
		assert.Error(t, err)
	})
}

func TestTextParser_ParseJQ(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	jsonInput := []byte(`{
		"users": [
			{"id": 1, "name": "Alice", "active": true},
			{"id": 2, "name": "Bob", "active": false},
			{"id": 3, "name": "Charlie", "active": true}
		]
	}`)

	t.Run("simple_filter", func(t *testing.T) {
		query := `.users[] | select(.active == true) | .name`
		result, err := parser.Parse("jq", jsonInput, nil, query)
		require.NoError(t, err)

		// Expecting a list of results
		assert.IsType(t, []any{}, result)
		list := result.([]any)
		assert.Contains(t, list, "Alice")
		assert.Contains(t, list, "Charlie")
		assert.Len(t, list, 2)
	})

	t.Run("object_construction", func(t *testing.T) {
		query := `{active_users: [.users[] | select(.active) | .name]}`
		result, err := parser.Parse("jq", jsonInput, nil, query)
		require.NoError(t, err)

		resMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.IsType(t, []any{}, resMap["active_users"])
		list := resMap["active_users"].([]any)
		assert.Equal(t, "Alice", list[0])
		assert.Equal(t, "Charlie", list[1])
	})

	t.Run("empty_query", func(t *testing.T) {
		_, err := parser.Parse("jq", jsonInput, nil, "")
		require.Error(t, err)
	})

	t.Run("invalid_query", func(t *testing.T) {
		_, err := parser.Parse("jq", jsonInput, nil, ".users[")
		require.Error(t, err)
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := parser.Parse("jq", []byte("{invalid"), nil, ".")
		require.Error(t, err)
	})
}
