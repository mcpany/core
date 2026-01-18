// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiagnostics_UnreachableUpstream verifies that LoadServices fails with a diagnostic error
// when a configured upstream service is unreachable.
func TestDiagnostics_UnreachableUpstream(t *testing.T) {
	// Pick a random port that is likely closed.
	// We can bind a listener to find a free port, then close it to ensure it's closed.
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	content := fmt.Sprintf(`
upstream_services: {
	name: "unreachable-service"
	http_service: {
		address: "http://localhost:%d"
	}
}
`, port)

	filePath := createTempConfigFile(t, content)
	fs := afero.NewOsFs()
	fileStore := NewFileStore(fs, []string{filePath})

	_, err = LoadServices(context.Background(), fileStore, "server")

	// Assert that we get an error about connection refused or timeout
	require.Error(t, err, "Expected LoadServices to fail due to unreachable upstream")
	// The error might be "connection refused" or "i/o timeout" or "context deadline exceeded" depending on OS/timing
	assert.True(t,
		assert.Contains(t, err.Error(), "connection refused") ||
		assert.Contains(t, err.Error(), "timeout") ||
		assert.Contains(t, err.Error(), "deadline exceeded"),
		"Error should indicate connection issue or timeout",
	)
	assert.Contains(t, err.Error(), fmt.Sprintf("localhost:%d", port), "Error should mention the target address")
}
