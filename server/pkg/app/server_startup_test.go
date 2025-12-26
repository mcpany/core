package app

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupReliability_ServiceFailure(t *testing.T) {
	// This test simulates a scenario where one of the upstream services is down or misconfigured.
	// We want to ensure that MCP Any still starts up and serves the other valid services.
	// Current behavior: The server crashes/fails to start if a sub-service fails.
	// Desired behavior: The server starts up, logging the error for the failed service.

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	configContent := `
upstream_services:
 - name: "valid-service"
   http_service:
     address: "http://localhost:8080"
     tools:
       - name: "echo"
         call_id: "echo_call"
     calls:
       echo_call:
         id: "echo_call"
         endpoint_path: "/echo"
         method: "HTTP_METHOD_POST"
 - name: "failing-service"
   http_service:
     address: "::invalid-url::" # This should likely cause an error in NewUpstream or Register
     tools: []
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, fmt.Sprintf("localhost:%d", port), "", []string{"/config.yaml"}, 1*time.Second)
	}()

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Check if app is still running (channel empty)
	select {
	case err := <-errChan:
		// If it returned, it means it crashed or stopped.
		if err != nil {
			t.Fatalf("Server crashed: %v", err)
		}
	default:
		// Still running
	}

	// Check if valid-service is registered
	_, ok := app.ToolManager.GetTool("valid-service.echo")

	assert.True(t, ok, "valid-service should be registered")

	cancel()
	<-errChan
}
