package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYamlEngine_Unmarshal_Error(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create invalid YAML file (tab indentation is invalid in YAML)
	require.NoError(t, afero.WriteFile(fs, "/config/invalid.yaml", []byte("\tinvalid: yaml"), 0644))

	store := NewFileStore(fs, []string{"/config/invalid.yaml"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal YAML")
}

func TestUnmarshal_NaN(t *testing.T) {
	engine := &yamlEngine{}
	// YAML supports NaN, but JSON does not.
	// This should fail at json.Marshal step.
	yamlData := []byte("global_settings:\n  log_level: .nan")

	// Wait, log_level is STRING in proto.
	// yaml.Unmarshal might decode .nan as string ".nan" if field is string?
	// But yaml.Unmarshal unmarshals into INTERFACE map first.
	// So it detects type. .nan is float in YAML.
	// So map["global_settings"]["log_level"] = NaN (float64).
	// json.Marshal will fail.

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	err := engine.Unmarshal(yamlData, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal map to JSON")
}

func TestMultiStore_HasConfigSources(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Case 1: Empty MultiStore
	ms := NewMultiStore()
	assert.False(t, ms.HasConfigSources())

	// Case 2: Store with sources
	s1 := NewFileStore(fs, []string{"/config"})
	ms = NewMultiStore(s1)
	assert.True(t, ms.HasConfigSources())

	// Case 3: Store without sources
	s2 := NewFileStore(fs, []string{})
	ms = NewMultiStore(s2)
	assert.False(t, ms.HasConfigSources())

	// Case 4: Mixed
	ms = NewMultiStore(s2, s1)
	assert.True(t, ms.HasConfigSources())
}

func TestMultiStore_Load(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/1.yaml", []byte(`
global_settings:
  log_level: INFO
`), 0644))
	require.NoError(t, afero.WriteFile(fs, "/config/2.yaml", []byte(`
global_settings:
  api_key: "my-key"
`), 0644))

	s1 := NewFileStore(fs, []string{"/config/1.yaml"})
	s2 := NewFileStore(fs, []string{"/config/2.yaml"})
	ms := NewMultiStore(s1, s2)

	cfg, err := ms.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())
	assert.Equal(t, "my-key", cfg.GetGlobalSettings().GetApiKey())
}
