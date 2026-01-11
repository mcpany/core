package diagnostic

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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

	// Case 1: Port is free
	err = CheckPortAvailability(address)
	assert.NoError(t, err)

	// Case 2: Port is taken
	l, err = net.Listen("tcp", address)
	assert.NoError(t, err)
	defer l.Close()

	err = CheckPortAvailability(address)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port binding check failed")
}
