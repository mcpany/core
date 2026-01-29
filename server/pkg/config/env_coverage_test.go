// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestConvertKind(t *testing.T) {
	// Bool - Valid
	res := convertKind(protoreflect.BoolKind, "true")
	assert.Equal(t, true, res)

	// Bool - Invalid (fallback to string)
	res = convertKind(protoreflect.BoolKind, "invalid")
	assert.Equal(t, "invalid", res)

	// String
	res = convertKind(protoreflect.StringKind, "foo")
	assert.Equal(t, "foo", res)

	// Int (currently returns string)
	res = convertKind(protoreflect.Int32Kind, "123")
	assert.Equal(t, "123", res)
}

func TestApplyEnvVars_Complex(t *testing.T) {
	// Test repeated message field with index
	// MCPANY__UPSTREAM_SERVICES__0__NAME = "s1"

	environ := []string{
		"MCPANY__UPSTREAM_SERVICES__0__NAME=s1",
		"MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS__0=127.0.0.1",
		"MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS__1=10.0.0.1",
		"MCPANY__GLOBAL_SETTINGS__DEBUGGER__ENABLED=true",
	}

	m := make(map[string]interface{})
	cfg := configv1.McpAnyServerConfig_builder{}.Build()

	applyEnvVarsFromSlice(m, environ, cfg)

	// Verify map structure
	services, ok := m["upstream_services"].(map[string]interface{})
	assert.True(t, ok)
	svc0, ok := services["0"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "s1", svc0["name"])

	gs, ok := m["global_settings"].(map[string]interface{})
	assert.True(t, ok)

	ips, ok := gs["allowed_ips"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "127.0.0.1", ips["0"])
	assert.Equal(t, "10.0.0.1", ips["1"])

	debugger, ok := gs["debugger"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, debugger["enabled"])
}

func TestResolveEnvValue_List_Mismatch(t *testing.T) {
	// MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS__0__SUBFIELD = val
	// allowed_ips is scalar string, cannot have subfield.

	environ := []string{
		"MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS__0__SUBFIELD=val",
	}
	m := make(map[string]interface{})
	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	applyEnvVarsFromSlice(m, environ, cfg)

	// Should result in: allowed_ips: { "0": { "subfield": "val" } } in map

	gs, ok := m["global_settings"].(map[string]interface{})
	assert.True(t, ok)
	ips, ok := gs["allowed_ips"].(map[string]interface{})
	assert.True(t, ok)
	idx0, ok := ips["0"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "val", idx0["subfield"])
}

func TestResolveEnvValue_CSV(t *testing.T) {
	// MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS=1.1.1.1,2.2.2.2
	environ := []string{
		"MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS=1.1.1.1,2.2.2.2",
	}
	m := make(map[string]interface{})
	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	applyEnvVarsFromSlice(m, environ, cfg)

	gs, ok := m["global_settings"].(map[string]interface{})
	assert.True(t, ok)
	val := gs["allowed_ips"]
	list, ok := val.([]interface{})
	assert.True(t, ok)
	if assert.Len(t, list, 2) {
		assert.Equal(t, "1.1.1.1", list[0])
		assert.Equal(t, "2.2.2.2", list[1])
	}
}

func TestFixTypes(t *testing.T) {
	// Manually construct a map that mimics Env Var map structure (indices as keys)
	// and run fixTypes to convert to slices.

	m := map[string]interface{}{
		"global_settings": map[string]interface{}{
			"allowed_ips": map[string]interface{}{
				"0": "1.1.1.1",
				"1": "2.2.2.2",
			},
		},
		"upstream_services": map[string]interface{}{
			"0": map[string]interface{}{
				"name": "s1",
			},
		},
	}

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	fixTypes(m, cfg.ProtoReflect().Descriptor())

	// allowed_ips should be []interface{}
	gs, ok := m["global_settings"].(map[string]interface{})
	assert.True(t, ok)
	ips, ok := gs["allowed_ips"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, ips, 2)

	// upstream_services should be []interface{}
	svcs, ok := m["upstream_services"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, svcs, 1)
}
