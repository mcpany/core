
package app

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_InvalidConfig_UnknownField(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a dummy config file with the bug.
	configContent := `
upstream_services:
 - name: "test-http-service"
   http_service:
     address: "http://localhost:8080"
     tools:
       - schema:
           name: "echo"
         call_id: "echo_call"
     calls:
       echo_call:
         id: "echo_call"
         endpoint_path: "/echo"
         method: "HTTP_METHOD_POST"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	err = app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, 5*time.Second)

	require.Error(t, err, "app.Run should return an error due to the invalid config")
	assert.Contains(t, err.Error(), "unknown field \"schema\"")
}
