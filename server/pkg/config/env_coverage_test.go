package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
    "github.com/stretchr/testify/assert"
)

func TestApplyEnvVarsFromSlice_Coverage(t *testing.T) {
	// We use a map representing the config structure
	m := make(map[string]interface{})

	// We need a proto message to guide type resolution
	v := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{},
	}

	env := []string{
		// Boolean field
		"MCPANY__GLOBAL_SETTINGS__USE_SUDO_FOR_DOCKER=true",
        // String field
        "MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS=:50050",
		// List of strings (repeated string)
		"MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS=127.0.0.1,10.0.0.1",
        // List of strings with spaces
        "MCPANY__GLOBAL_SETTINGS__ALLOWED_FILE_PATHS=/tmp, /var",
        // Ignore non-MCPANY
        "OTHER_VAR=value",
        // Ignore malformed
        "MCPANY__MALFORMED",
        // Invalid boolean
        "MCPANY__GLOBAL_SETTINGS__DLP__ENABLED=not-a-bool",
	}

	applyEnvVarsFromSlice(m, env, v)

	// Verify map content
	gs, ok := m["global_settings"].(map[string]interface{})
	assert.True(t, ok)

	// Boolean should be bool type
	assert.Equal(t, true, gs["use_sudo_for_docker"])

	// String
    assert.Equal(t, ":50050", gs["mcp_listen_address"])

	// List of strings
	ips, ok := gs["allowed_ips"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "127.0.0.1", ips[0])
	assert.Equal(t, "10.0.0.1", ips[1])

    // List of strings with spaces
    paths, ok := gs["allowed_file_paths"].([]interface{})
    assert.True(t, ok)
    assert.Equal(t, "/tmp", paths[0])
    assert.Equal(t, "/var", paths[1])

    // Invalid boolean should remain string
    dlp, ok := gs["dlp"].(map[string]interface{})
    assert.True(t, ok)
    assert.Equal(t, "not-a-bool", dlp["enabled"])
}

func TestConvertKind_Coverage(t *testing.T) {
    // We can't call convertKind directly as it is unexported and we are in the same package but let's see if we can trigger other paths via applyEnvVarsFromSlice

    // We already triggered BoolKind.

    // Trigger "path mismatch" or deep nesting
    m := make(map[string]interface{})
    v := &configv1.McpAnyServerConfig{}

    env := []string{
        // Deep nesting that needs map creation
        "MCPANY__GLOBAL_SETTINGS__AUDIT__ENABLED=true",
    }

    applyEnvVarsFromSlice(m, env, v)

    gs, ok := m["global_settings"].(map[string]interface{})
    assert.True(t, ok)
    audit, ok := gs["audit"].(map[string]interface{})
    assert.True(t, ok)
    assert.Equal(t, true, audit["enabled"])
}

func TestResolveEnvValue_Coverage(t *testing.T) {
     // Test list indexing if supported by logic?
     // logic: if currentFd != nil && currentFd.IsList() ...

     // configv1.UpstreamServiceConfig repeated UpstreamServiceConfig upstream_services = 2;

     m := make(map[string]interface{})
     v := &configv1.McpAnyServerConfig{}

     env := []string{
         // Accessing list by index?? logic seems to support it?
         // parts := strings.SplitN(s, ":", 2) in expand, but here we are in applyEnvVars
         // applyEnvVars splits by __

         // resolveEnvValue iterates path.
         // if currentFd.IsList() -> part is index.

         // Let's try to set a name of the first service
         "MCPANY__UPSTREAM_SERVICES__0__NAME=myservice",
     }

     applyEnvVarsFromSlice(m, env, v)

     // Check structure
     // upstream_services should be a slice of maps?
     // fixTypes converts map to slice.
     // But applyEnvVars creates maps.

     // m["upstream_services"]["0"]["name"] = "myservice"

     us, ok := m["upstream_services"].(map[string]interface{})
     assert.True(t, ok)
     svc0, ok := us["0"].(map[string]interface{})
     assert.True(t, ok)
     assert.Equal(t, "myservice", svc0["name"])

     // Now run fixTypes on it?
     // applyEnvVars calls fixTypes(m, v.ProtoReflect().Descriptor())
     // So m["upstream_services"] should be converted to slice!

     // Wait, fixTypes modifies m in place?
     // applyEnvVars:
     // 	applyEnvVarsFromSlice(m, os.Environ(), v)
	 // if v != nil {
	 // 	fixTypes(m, v.ProtoReflect().Descriptor())
	 // }

     // My test TestApplyEnvVarsFromSlice_Coverage calls applyEnvVarsFromSlice directly, skipping fixTypes.
     // So I should call fixTypes manually to test it.

     fixTypes(m, v.ProtoReflect().Descriptor())

     usSlice, ok := m["upstream_services"].([]interface{})
     assert.True(t, ok)
     assert.Len(t, usSlice, 1)
     svc0Map, ok := usSlice[0].(map[string]interface{})
     assert.True(t, ok)
     assert.Equal(t, "myservice", svc0Map["name"])
}

func TestResolveEnvValue_Mistmatch_Coverage(t *testing.T) {
    m := make(map[string]interface{})
    v := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{},
    }

    env := []string{
        // Path continues but field is not a message (Scalar)
        // mcp_listen_address is scalar string.
        "MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS__SUBFIELD=val",

        // List element access on scalar list, but trying to access subfield
        "MCPANY__GLOBAL_SETTINGS__ALLOWED_IPS__0__SUBFIELD=val",
    }

    applyEnvVarsFromSlice(m, env, v)

    gs, ok := m["global_settings"].(map[string]interface{})
    assert.True(t, ok)
    // Should contain resolved values as strings because resolution failed
    // mcp_listen_address will be a map because applyEnvVars forced creation of map structure,
    // but resolveEnvValue returned "val" for the leaf?

    // applyEnvVars:
    // current[section] = next (map)
    // ...
    // current[section] = resolvedValue

    // So global_settings.mcp_listen_address becomes a map containing subfield=val?
    // Because resolveEnvValue returns "val" (fallback).

    addrMap, ok := gs["mcp_listen_address"].(map[string]interface{})
    assert.True(t, ok)
    assert.Equal(t, "val", addrMap["subfield"])

    ipsMap, ok := gs["allowed_ips"].(map[string]interface{})
    assert.True(t, ok)
    zeroMap, ok := ipsMap["0"].(map[string]interface{})
    assert.True(t, ok)
    assert.Equal(t, "val", zeroMap["subfield"])
}
