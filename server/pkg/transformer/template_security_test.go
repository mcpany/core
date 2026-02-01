// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextTemplate_XMLInjection(t *testing.T) {
	t.Parallel()
	// XML template
	templateString := `<user><name>{{name}}</name></user>`
	tpl, err := NewTemplate(templateString, "{{", "}}")
	require.NoError(t, err)

	// Malicious input
	params := map[string]any{
		"name": `foo</name><role>admin</role><name>bar`,
	}

	rendered, err := tpl.Render(params)
	require.NoError(t, err)

	// Expect proper XML escaping
	expectedSafe := `<user><name>foo&lt;/name&gt;&lt;role&gt;admin&lt;/role&gt;&lt;name&gt;bar</name></user>`
	assert.Equal(t, expectedSafe, rendered)
}
