// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformer_JSONInput(t *testing.T) {
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
		"email":        `{.person.contacts[?(@.type=="email")].value}`,
	}
	parsedData, err := parser.Parse("json", jsonInput, config)
	require.NoError(t, err)

	templateString := "Name: {{name}}, Email: {{email}}"
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	rendered, err := tpl.Render(parsedData)
	require.NoError(t, err)
	assert.Equal(t, "Name: test, Email: test@example.com", rendered)
}

func TestTransformer_XMLInput(t *testing.T) {
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
	parsedData, err := parser.Parse("xml", xmlInput, config)
	require.NoError(t, err)

	templateString := "Cell 1: {{cell1}}, Cell 2: {{cell2}}"
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	rendered, err := tpl.Render(parsedData)
	require.NoError(t, err)
	assert.Equal(t, "Cell 1: Apples, Cell 2: Bananas", rendered)
}

func TestTransformer_TextInput(t *testing.T) {
	parser := NewTextParser()
	textInput := []byte(`Event: "user_login", Status: "success", User-ID: "user-123@example.com"`)
	config := map[string]string{
		"event":  `Event: "([^"]+)"`,
		"email":  `User-ID: "([^"]+)"`,
	}
	parsedData, err := parser.Parse("text", textInput, config)
	require.NoError(t, err)

	templateString := "Event: {{event}}, User: {{email}}"
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	rendered, err := tpl.Render(parsedData)
	require.NoError(t, err)
	assert.Equal(t, "Event: user_login, User: user-123@example.com", rendered)
}
