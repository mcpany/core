// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTextParser_ParseJSON_MultipleMatches verifies the behavior when a JSONPath selector
// matches multiple elements.
//
// CURRENT LIMITATION: The parser currently returns only the FIRST match, even if multiple
// elements are selected. This test documents this behavior. If this behavior changes to
// return a list, this test should be updated.
func TestTextParser_ParseJSON_MultipleMatches(t *testing.T) {
	parser := NewTextParser()
	jsonInput := []byte(`{
		"items": [
			{"id": 1},
			{"id": 2},
			{"id": 3}
		]
	}`)
	config := map[string]string{
		"ids": `{.items[*].id}`,
	}
	result, err := parser.Parse("json", jsonInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)

	// Expecting only the FIRST ID (float64(1)) due to current implementation
	assert.Equal(t, float64(1), resMap["ids"])
}

// TestTextParser_ParseXML_MultipleMatches verifies the behavior when an XPath selector
// matches multiple elements.
//
// CURRENT LIMITATION: The parser currently returns only the FIRST match.
func TestTextParser_ParseXML_MultipleMatches(t *testing.T) {
	parser := NewTextParser()
	xmlInput := []byte(`
		<root>
			<item><id>1</id></item>
			<item><id>2</id></item>
			<item><id>3</id></item>
		</root>
	`)
	config := map[string]string{
		"ids": `//item/id`,
	}
	result, err := parser.Parse("xml", xmlInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)

	// Expecting only the FIRST ID "1" due to current implementation
	assert.Equal(t, "1", resMap["ids"])
}

// TestTextParser_ParseText_MultipleMatches verifies the behavior when a Regex selector
// matches multiple times.
//
// CURRENT LIMITATION: The parser currently returns only the FIRST match.
func TestTextParser_ParseText_MultipleMatches(t *testing.T) {
	parser := NewTextParser()
	textInput := []byte(`ID: 1, ID: 2, ID: 3`)
	config := map[string]string{
		"ids": `ID: (\d+)`,
	}
	result, err := parser.Parse("text", textInput, config, "")
	require.NoError(t, err)
	resMap := result.(map[string]any)

	// Expecting only the FIRST ID "1" due to current implementation
	assert.Equal(t, "1", resMap["ids"])
}
