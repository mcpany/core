/*
 * Copyright 2025 Author(s) of MCPX
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

func TestTransformer_ParseAndRender(t *testing.T) {
	// 1. Setup the parser
	parser := NewTextParser()
	jsonInput := []byte(`{"data":{"weather":{"description":"clear sky","temperature":25}}}`)
	config := map[string]string{
		"description": `{.data.weather.description}`,
		"temperature": `{.data.weather.temperature}`,
	}

	// 2. Parse the input
	parsedResult, err := parser.Parse("json", jsonInput, config)
	require.NoError(t, err)
	assert.Equal(t, "clear sky", parsedResult["description"])
	assert.Equal(t, float64(25), parsedResult["temperature"])

	// 3. Setup the template
	template, err := NewTextTemplate("The weather is {{.description}} with a temperature of {{.temperature}} degrees.")
	require.NoError(t, err)

	// 4. Render the output
	renderedOutput, err := template.Render(parsedResult)
	require.NoError(t, err)

	// 5. Assert the final output
	expectedOutput := "The weather is clear sky with a temperature of 25 degrees."
	assert.Equal(t, expectedOutput, renderedOutput)
}
