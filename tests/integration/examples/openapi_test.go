//go:build e2e

package examples

import (
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/mcpxy/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	cmd := exec.Command("/bin/bash", "start.sh")
	cmd.Dir = root + "/examples/upstream/openapi"
	err = cmd.Start()
	require.NoError(t, err)
	defer cmd.Process.Kill()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:8080")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, "Hello, world!", string(body))
}
