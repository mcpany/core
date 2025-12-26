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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func TestRun_GrpcReflectionFailure_Retry(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find a free port for the gRPC service
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close() // Close it so it fails initially
	grpcAddr := fmt.Sprintf("localhost:%d", grpcPort)

	// Create a config with a gRPC service using reflection
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "grpc-service"
    grpc_service:
      address: "%s"
      use_reflection: true
      tools:
        - name: "test-tool"
          call_id: "test-call"
      calls:
        test-call:
          service: "test.TestService"
          method: "Test"
`, grpcAddr)
	err = afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)

	// Start the app
	go func() {
		// Use ephemeral ports for the app itself
		errChan <- app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, 5*time.Second)
	}()

	// Wait a bit for initial registration attempt to fail
	time.Sleep(2 * time.Second)

	// Verify service is NOT registered yet
	_, ok := app.ToolManager.GetTool("grpc-service.test-tool")
	assert.False(t, ok, "grpc-service.test-tool should NOT be registered yet")

	// Now start the gRPC server
	lis, err := net.Listen("tcp", grpcAddr)
	require.NoError(t, err)
	s := grpc.NewServer()
	// We need to register reflection service
	reflection.Register(s)

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	// Wait for retry to happen.
	// We check if the tool appears. It MIGHT fail tool registration because method is missing,
	// BUT `reflection` itself should succeed, so `ParseProtoByReflection` succeeds.
	// If `ParseProtoByReflection` succeeds, it proceeds to register tools.
	// It logs errors for missing methods but might register "ServerReflectionInfo" tool if auto-discovery is on?
	// Ah, the logs showed "Registered gRPC tool" tool_id=ServerReflectionInfo.
	// So we CAN check for "grpc-service.ServerReflectionInfo".

	require.Eventually(t, func() bool {
		_, ok := app.ToolManager.GetTool("grpc-service.ServerReflectionInfo")
		return ok
	}, 10*time.Second, 500*time.Millisecond, "grpc-service.ServerReflectionInfo should be registered after retry")

	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
