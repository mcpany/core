package app

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortConflict(t *testing.T) {
	// 1. Start a listener on a random port
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer l.Close()

	portStr := l.Addr().String()

	// 2. Try to start the application on the same port
	app := NewApplication()
	opts := RunOptions{
		Ctx:             context.Background(),
		Fs:              afero.NewMemMapFs(),
		Stdio:           false,
		JSONRPCPort:     portStr, // Use the conflicting port
		GRPCPort:        "",
		ShutdownTimeout: 100 * time.Millisecond,
	}

	// We expect Run to fail
	err = app.Run(opts)
	require.Error(t, err)

	// 3. Verify the error message is helpful
	// Current expected behavior: generic bind error
	// Desired behavior: helpful message
	t.Logf("Error received: %v", err)

	// Check if it contains the helpful hint
	assert.Contains(t, err.Error(), "address already in use")
	assert.Contains(t, err.Error(), "Tip: The port is already in use")
	assert.Contains(t, err.Error(), "--json-rpc-port")
}
