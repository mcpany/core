package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestApplyEnvVars_Types(t *testing.T) {
	// Setup env vars
	os.Setenv("MCPANY__GLOBAL_SETTINGS__USE_SUDO_FOR_DOCKER", "true")
	os.Setenv("MCPANY__GLOBAL_SETTINGS__LOG_LEVEL", "LOG_LEVEL_DEBUG")
    // Test invalid bool
    os.Setenv("MCPANY__GLOBAL_SETTINGS__MESSAGE_BUS__REDIS__ENABLE_TLS", "not-a-bool")

    defer os.Unsetenv("MCPANY__GLOBAL_SETTINGS__USE_SUDO_FOR_DOCKER")
    defer os.Unsetenv("MCPANY__GLOBAL_SETTINGS__LOG_LEVEL")
    defer os.Unsetenv("MCPANY__GLOBAL_SETTINGS__MESSAGE_BUS__REDIS__ENABLE_TLS")

	// We need a proto message to guide the type conversion
	cfg := &configv1.McpAnyServerConfig{}

	// Create a map that mimics YAML structure
	m := make(map[string]interface{})

	// Apply env vars
	applyEnvVars(m, cfg)

	// Check results
	gs, ok := m["global_settings"].(map[string]interface{})
	assert.True(t, ok)

	// Bool conversion
	assert.Equal(t, true, gs["use_sudo_for_docker"])

    // Enum/String (no conversion needed really, just string)
    assert.Equal(t, "LOG_LEVEL_DEBUG", gs["log_level"])

    // Invalid bool should be kept as string
    mb, ok := gs["message_bus"].(map[string]interface{})
    assert.True(t, ok)
    redis, ok := mb["redis"].(map[string]interface{})
    assert.True(t, ok)
    assert.Equal(t, "not-a-bool", redis["enable_tls"])
}

func TestApplyEnvVars_Lists(t *testing.T) {
    os.Setenv("MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS", "127.0.0.1,192.168.1.1")
    defer os.Unsetenv("MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS")

    cfg := &configv1.McpAnyServerConfig{}
    m := make(map[string]interface{})

    applyEnvVars(m, cfg)

    gs, ok := m["global_settings"].(map[string]interface{})
    assert.True(t, ok)

    ips, ok := gs["allowed_ips"].([]interface{})
    assert.True(t, ok)
    assert.Len(t, ips, 2)
    assert.Equal(t, "127.0.0.1", ips[0])
    assert.Equal(t, "192.168.1.1", ips[1])
}

func TestApplyEnvVars_BoolList(t *testing.T) {
    // There might not be a bool list in the config, but we can try if there is one.
    // Assuming no bool list exists, this might be hard to test properly using the real config struct.
}
