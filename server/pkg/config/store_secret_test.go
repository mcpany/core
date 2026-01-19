package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYamlEngine_SimplifiedSecret(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with simplified secret syntax (string instead of object)
	// Case 1: Map (env)
	// Case 2: Direct field (upstream_auth.api_key.value)
	require.NoError(t, afero.WriteFile(fs, "/config/simple_secret.yaml", []byte(`
upstream_services:
  - name: "simple-secret-test"
    mcp_service:
      stdio_connection:
        command: "env"
        env:
          API_KEY: "simple-value-in-map"
    upstream_auth:
      api_key:
        param_name: "X-API-Key"
        value: "simple-value-direct"
`), 0644))

	store := NewFileStore(fs, []string{"/config/simple_secret.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	svc := cfg.GetUpstreamServices()[0]

	// Verify Map Case
	stdio := svc.GetMcpService().GetStdioConnection()
	require.NotNil(t, stdio)
	envKey := stdio.GetEnv()["API_KEY"]
	require.NotNil(t, envKey)
	assert.Equal(t, "simple-value-in-map", envKey.GetPlainText())

	// Verify Direct Field Case
	auth := svc.GetUpstreamAuth()
	require.NotNil(t, auth)
	apiKey := auth.GetApiKey()
	require.NotNil(t, apiKey)
	secretVal := apiKey.GetValue()
	require.NotNil(t, secretVal)
	assert.Equal(t, "simple-value-direct", secretVal.GetPlainText())
}
