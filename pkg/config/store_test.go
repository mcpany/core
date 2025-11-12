
package config

import (
	"encoding/json"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

func TestNewEngine(t *testing.T) {
	t.Run("UnsupportedExtension", func(t *testing.T) {
		_, err := NewEngine("config.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported config file extension")
	})

	t.Run("JSONExtension", func(t *testing.T) {
		engine, err := NewEngine("config.json")
		assert.NoError(t, err)
		assert.IsType(t, &jsonEngine{}, engine)
	})
}

func TestJsonEngine_Unmarshal(t *testing.T) {
	engine := &jsonEngine{}

	t.Run("ValidJSON", func(t *testing.T) {
		validJSON := []byte(`{
			"global_settings": {
				"bind_address": "0.0.0.0:8080",
				"log_level": "INFO"
			}
		}`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(validJSON, cfg)
		require.NoError(t, err)
		assert.Equal(t, "0.0.0.0:8080", cfg.GetGlobalSettings().GetBindAddress())
		assert.Equal(t, configv1.GlobalSettings_INFO, cfg.GetGlobalSettings().GetLogLevel())
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{
			"global_settings": {
				"bind_address": "0.0.0.0:8080",
				"log_level": "INFO",
			}
		}`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(invalidJSON, cfg)
		assert.Error(t, err)
	})
}

func TestYamlEngine_Unmarshal(t *testing.T) {
	engine := &yamlEngine{}

	t.Run("InvalidYAML", func(t *testing.T) {
		invalidYAML := []byte(`
global_settings:
  bind_address: "0.0.0.0:8080"
  log_level: "INFO"
  protoc_version: "3.19.4"
- this is not valid
`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(invalidYAML, cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal YAML")
	})

	t.Run("ValidYAML", func(t *testing.T) {
		validYAML := []byte(`
global_settings:
  bind_address: "0.0.0.0:8080"
  log_level: "INFO"
`)
		cfg := &configv1.McpAnyServerConfig{}
		err := engine.Unmarshal(validYAML, cfg)
		require.NoError(t, err)
		assert.Equal(t, "0.0.0.0:8080", cfg.GetGlobalSettings().GetBindAddress())
		assert.Equal(t, configv1.GlobalSettings_INFO, cfg.GetGlobalSettings().GetLogLevel())
	})
}

// marshalError is a helper type that always returns an error when marshaled to JSON.
type marshalError struct{}

func (m *marshalError) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal error")
}

func (e *yamlEngine) UnmarshalWithFailingJSON(b []byte, v proto.Message) error {
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(b, &yamlMap); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	// Purposely cause a marshaling error by using a type that fails to marshal.
	_, err := json.Marshal(&marshalError{})
	return fmt.Errorf("failed to marshal map to JSON: %w", err)
}

func TestYamlEngine_Unmarshal_MarshalError(t *testing.T) {
	engine := &yamlEngine{}
	validYAML := []byte(`
global_settings:
  bind_address: "0.0.0.0:8080"
`)
	err := engine.UnmarshalWithFailingJSON(validYAML, &configv1.McpAnyServerConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal map to JSON")
}

func TestFileStore_Load(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Setup directory and config files
	require.NoError(t, fs.MkdirAll("configs/subdir", 0o755))
	afero.WriteFile(fs, "configs/01_base.yaml", []byte(`
global_settings:
  bind_address: "0.0.0.0:8080"
  log_level: "INFO"
upstream_services:
- id: "service-1"
  name: "first-service"
`), 0o644)

	afero.WriteFile(fs, "configs/02_override.yaml", []byte(`
global_settings:
  bind_address: "127.0.0.1:9090"
upstream_services:
- id: "service-2"
  name: "second-service"
`), 0o644)

	afero.WriteFile(fs, "configs/invalid.txt", []byte("invalid content"), 0o644)
	afero.WriteFile(fs, "malformed.yaml", []byte("bad-yaml:"), 0o644)
	require.NoError(t, fs.Mkdir("configs/subdir/empty", 0o755))

	testCases := []struct {
		name          string
		paths         []string
		expectErr     bool
		expectedCfg   *configv1.McpAnyServerConfig
		checkResult   func(t *testing.T, cfg *configv1.McpAnyServerConfig)
		expectedErrFn func(t *testing.T, err error)
	}{
		{
			name:  "Load single file",
			paths: []string{"configs/01_base.yaml"},
			checkResult: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				assert.Equal(t, "0.0.0.0:8080", cfg.GetGlobalSettings().GetBindAddress())
				assert.Len(t, cfg.GetUpstreamServices(), 1)
			},
		},
		{
			name:  "Load and merge multiple files",
			paths: []string{"configs/01_base.yaml", "configs/02_override.yaml"},
			checkResult: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				// Last one wins for scalar fields
				assert.Equal(t, "127.0.0.1:9090", cfg.GetGlobalSettings().GetBindAddress())
				// Repeated fields are appended
				assert.Len(t, cfg.GetUpstreamServices(), 2)
				assert.Equal(t, "service-1", cfg.GetUpstreamServices()[0].GetId())
				assert.Equal(t, "service-2", cfg.GetUpstreamServices()[1].GetId())
			},
		},
		{
			name:      "Path does not exist",
			paths:     []string{"nonexistent/"},
			expectErr: true,
			expectedErrFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to stat path")
			},
		},
		{
			name:      "Load with malformed file",
			paths:     []string{"malformed.yaml"},
			expectErr: true,
			expectedErrFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to unmarshal config")
			},
		},
		{
			name:  "Empty directory results in nil config",
			paths: []string{"configs/subdir/empty"},
			checkResult: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				assert.Nil(t, cfg)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewFileStore(fs, tc.paths)
			cfg, err := store.Load()

			if tc.expectErr {
				require.Error(t, err)
				if tc.expectedErrFn != nil {
					tc.expectedErrFn(t, err)
				}
			} else {
				require.NoError(t, err)
				if tc.checkResult != nil {
					tc.checkResult(t, cfg)
				}
			}
		})
	}
}
