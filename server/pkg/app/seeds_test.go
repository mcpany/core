// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltinTemplates_Integrity(t *testing.T) {
	assert.NotEmpty(t, BuiltinTemplates, "BuiltinTemplates should not be empty")

	for _, tmpl := range BuiltinTemplates {
		t.Run(tmpl.GetName(), func(t *testing.T) {
			assert.NotEmpty(t, tmpl.GetId())
			assert.NotEmpty(t, tmpl.GetName())

			// Verify Schema
			schema := tmpl.GetConfigurationSchema()
			assert.NotEmpty(t, schema)
			var schemaObj map[string]any
			err := json.Unmarshal([]byte(schema), &schemaObj)
			assert.NoError(t, err, "ConfigurationSchema should be valid JSON")

			// Verify Command
			cmdSvc := tmpl.GetCommandLineService()
			// Only CommandLineService is expected for current builtin templates?
			// Checking seeds.go, yes, all use mkTemplate which sets CommandLineService.
			// But future templates might differ. For now, asserting it's present if it's what mkTemplate does.
			// But maybe check if ANY service type is set?
			// The current seeds.go implementation ONLY supports CLI based templates via mkTemplate.
			// Except "memory" which has "npx -y @modelcontextprotocol/server-memory".

			assert.NotNil(t, cmdSvc)
			assert.NotEmpty(t, cmdSvc.GetCommand())

			// Verify AutoDiscover
			assert.True(t, tmpl.GetAutoDiscoverTool())
		})
	}
}

func TestMkTemplate(t *testing.T) {
	id := "test-id"
	name := "Test Name"
	schema := `{"type":"object"}`
	command := "echo hello"

	tmpl := mkTemplate(id, name, schema, command)

	assert.Equal(t, id, tmpl.GetId())
	assert.Equal(t, name, tmpl.GetName())
	assert.Equal(t, "1.0.0", tmpl.GetVersion())
	assert.Equal(t, schema, tmpl.GetConfigurationSchema())

	cmd := tmpl.GetCommandLineService()
	assert.NotNil(t, cmd)
	assert.Equal(t, command, cmd.GetCommand())
	assert.NotNil(t, cmd.GetEnv())

	assert.True(t, tmpl.GetAutoDiscoverTool())
}
