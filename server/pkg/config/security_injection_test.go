package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReproduceInjection(t *testing.T) {
	yamlContent := []byte(`
upstream_services:
  - name: "weather-api"
    http_service:
      address: "${WEATHER_API_URL}"
`)

    // Set env var with quotes that would break YAML string if naively expanded
    brokenValue := `https://example.com" broken: "yes`
    os.Setenv("WEATHER_API_URL", brokenValue)
    defer os.Unsetenv("WEATHER_API_URL")

	engine, err := NewEngine("config.yaml")
	require.NoError(t, err)

    // Unmarshal should succeed now because we expand safely in AST
	cfg := &configv1.McpAnyServerConfig{}
	err = engine.Unmarshal(yamlContent, cfg)

	assert.NoError(t, err)

    // Verify the value was injected as a string content, not syntax
    assert.Equal(t, brokenValue, cfg.UpstreamServices[0].GetHttpService().GetAddress())
}
