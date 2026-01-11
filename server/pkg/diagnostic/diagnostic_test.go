// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/spf13/afero"
)

func TestEnhanceConfigError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "Nil error",
			input:    nil,
			expected: "",
		},
		{
			name:     "Generic error",
			input:    errors.New("something went wrong"),
			expected: "something went wrong",
		},
		{
			name:     "Proto unknown field services",
			input:    fmt.Errorf(`proto: (line 1:2): unknown field "services"`),
			expected: "configuration error: unknown field 'services'. Did you mean 'upstream_services'?",
		},
		{
			name:     "Proto unknown field with trailing char",
			input:    fmt.Errorf(`proto: (line 1:2): unknown field "services")`),
			expected: "configuration error: unknown field 'services'. Did you mean 'upstream_services'?",
		},
		{
			name:     "Wrapped unknown field",
			input:    fmt.Errorf(`failed to load config: proto: (line 1:2): unknown field "mcpListenAddress"`),
			expected: "configuration error: unknown field 'mcpListenAddress'. Did you mean 'global_settings.mcp_listen_address'?",
		},
		{
			name:     "YAML syntax error",
			input:    errors.New("failed to unmarshal YAML: yaml: line 1: did not find expected ',' or ']'"),
			expected: "configuration syntax error: failed to unmarshal YAML: yaml: line 1: did not find expected ',' or ']'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnhanceConfigError(tt.input)
			if tt.input == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, err.Error())
			}
		})
	}
}

func TestCheckPortAvailability(t *testing.T) {
	// Find a random free port
	l, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	address := fmt.Sprintf("localhost:%d", port)
	ctx := context.Background()

	// Case 1: Port is free
	err = CheckPortAvailability(ctx, address)
	assert.NoError(t, err)

	// Case 2: Port is taken
	l, err = net.Listen("tcp", address)
	assert.NoError(t, err)
	defer l.Close()

	err = CheckPortAvailability(ctx, address)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port binding check failed")
}

func TestRunDiagnostics(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()

	// 1. Happy Path: No config file (valid), port available
	l, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	addr := fmt.Sprintf("localhost:%d", port)

	err = RunDiagnostics(ctx, fs, []string{}, addr)
	assert.NoError(t, err)

	// 2. Bad Config (Syntax Error)
	_ = afero.WriteFile(fs, "bad.yaml", []byte("bad: ["), 0644)
	err = RunDiagnostics(ctx, fs, []string{"bad.yaml"}, addr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration syntax error")

	// 3. Bad Config (Unknown Field)
	_ = afero.WriteFile(fs, "unknown.yaml", []byte("invalid: value"), 0644)
	err = RunDiagnostics(ctx, fs, []string{"unknown.yaml"}, addr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field 'invalid'")

	// 4. Port Conflict
	// Create valid config
	_ = afero.WriteFile(fs, "good.yaml", []byte("global_settings: {}"), 0644)

	l, err = net.Listen("tcp", addr)
	assert.NoError(t, err)
	defer l.Close()

	err = RunDiagnostics(ctx, fs, []string{"good.yaml"}, addr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network check failed")
}
