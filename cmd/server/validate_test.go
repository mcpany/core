/*
 * Copyright 2024 Author(s) of MCP Any
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

package main

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func execute(cmd *cobra.Command, args ...string) (string, string, error) {
	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	root := &cobra.Command{}
	root.AddCommand(cmd)
	root.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH")
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	root.SetArgs(append([]string{cmd.Use}, args...))

	err := root.Execute()

	return bufOut.String(), bufErr.String(), err
}

func TestValidateCommand(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create a valid config file
	validConfigFile, err := afero.TempFile(fs, "", "valid-*.yaml")
	require.NoError(t, err)
	defer validConfigFile.Close()
	_, err = validConfigFile.WriteString(`
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
`)
	require.NoError(t, err)

	// Create an invalid config file
	invalidConfigFile, err := afero.TempFile(fs, "", "invalid-*.yaml")
	require.NoError(t, err)
	defer invalidConfigFile.Close()
	_, err = invalidConfigFile.WriteString(`
upstreamServices:
  - name: "my-http-service"
`)
	require.NoError(t, err)

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  string
	}{
		{
			name:           "valid config",
			args:           []string{"--config-path", validConfigFile.Name()},
			expectedOutput: "Configuration is valid\n",
			expectedError:  "",
		},
		{
			name:           "invalid config",
			args:           []string{"--config-path", invalidConfigFile.Name()},
			expectedOutput: "",
			expectedError:  "configuration is invalid: config validation failed: service 'my-http-service': service has no service_config",
		},
		{
			name:           "no config path",
			args:           []string{},
			expectedOutput: "",
			expectedError:  "at least one configuration path must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newValidateCommand(fs)
			out, _, err := execute(cmd, tt.args...)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedOutput, out)
		})
	}
}
