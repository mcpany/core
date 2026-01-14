package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
    "github.com/stretchr/testify/assert"
)

func TestYamlEngine_Coverage(t *testing.T) {
    engine, err := NewEngine("config.yaml")
    assert.NoError(t, err)

    v := &configv1.McpAnyServerConfig{}

    // Malformed YAML
    err = engine.Unmarshal([]byte("key: : value"), v)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to unmarshal YAML")

    // Valid YAML but invalid Proto (e.g. unknown field)
    // Note: mcpServers is special case handled in code
    err = engine.Unmarshal([]byte("mcpServers: {}"), v)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "Did you mean \"upstream_services\"?")

    // Schema validation fail (structural)
    // upstream_services should be an array of objects
    err = engine.Unmarshal([]byte("upstream_services: not-an-array"), v)
    assert.Error(t, err)
    // Error message comes from protojson or jsonschema
}

func TestNewEngine_Coverage(t *testing.T) {
    _, err := NewEngine("config.unknown")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unsupported config file extension")

    _, err = NewEngine("config.json")
    assert.NoError(t, err)

    _, err = NewEngine("config.textproto")
    assert.NoError(t, err)
}
