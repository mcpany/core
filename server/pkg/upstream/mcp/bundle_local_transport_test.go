package mcp

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleLocalTransport_Connect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		transport := &BundleLocalTransport{
			Command: "echo",
			Args:    []string{"hello"},
			Env:     []string{"TEST=1"},
			WorkingDir: "/tmp",
		}

		conn, err := transport.Connect(ctx)
		require.NoError(t, err)
		assert.NotNil(t, conn)

		// It returns a stdio transport connection
		err = conn.Close()
		require.NoError(t, err)
	})

	t.Run("missing command", func(t *testing.T) {
		transport := &BundleLocalTransport{
			Command: "does-not-exist-command-12345",
			Args:    []string{"hello"},
		}

		_, err := transport.Connect(ctx)
		require.Error(t, err)
	})
}
