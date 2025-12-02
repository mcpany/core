/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestValidate(t *testing.T) {
	// Create temporary files for mTLS tests
	clientCertFile, err := os.CreateTemp("", "client.crt")
	require.NoError(t, err)
	defer os.Remove(clientCertFile.Name())

	clientKeyFile, err := os.CreateTemp("", "client.key")
	require.NoError(t, err)
	defer os.Remove(clientKeyFile.Name())

	caCertFile, err := os.CreateTemp("", "ca.crt")
	require.NoError(t, err)
	defer os.Remove(caCertFile.Name())

	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "invalid command line service - empty command",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "cmd-svc-1",
							"command_line_service": {
								"command": ""
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "cmd-svc-1": command_line_service has empty command`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(tt.config, Server)
			assert.Len(t, validationErrors, tt.expectedErrorCount)
			if tt.expectedErrorCount > 0 {
				require.NotEmpty(t, validationErrors)
				assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
			}
		})
	}
}
