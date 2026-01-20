package config

import (
	"os"
	"testing"
    "context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReproduceConfig(t *testing.T) {
	yamlContent := []byte(`
upstream_services:
  - name: "weather-api"
    http_service:
      address: "${WEATHER_API_URL}"
`)

    // unset env var to be sure
    os.Unsetenv("WEATHER_API_URL")

	engine, err := NewEngine("config.yaml")
	require.NoError(t, err)

    // Unmarshal should fail because env var is missing
    cfg := &configv1.McpAnyServerConfig{}
    err = engine.Unmarshal(yamlContent, cfg)
    assert.Error(t, err)
    t.Logf("Expected Unmarshal error (missing env): %v", err)

    // Now set it
    os.Setenv("WEATHER_API_URL", "https://api.weather.com")
    defer os.Unsetenv("WEATHER_API_URL")

    // Unmarshal should pass
    cfg = &configv1.McpAnyServerConfig{}
    err = engine.Unmarshal(yamlContent, cfg)
	assert.NoError(t, err)

    assert.Equal(t, "https://api.weather.com", cfg.UpstreamServices[0].GetHttpService().GetAddress())

    // Now validate
    validationErrors := Validate(context.Background(), cfg, Server)
    assert.Empty(t, validationErrors)
}
