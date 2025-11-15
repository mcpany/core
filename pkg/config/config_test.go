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
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDefineFlags(t *testing.T) {
	cmd := &cobra.Command{}
	DefineFlags(cmd)
	assert.NotNil(t, cmd.PersistentFlags().Lookup("mcp-listen-address"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config-path"))
	assert.NotNil(t, cmd.Flags().Lookup("grpc-port"))
	assert.NotNil(t, cmd.Flags().Lookup("stdio"))
	assert.NotNil(t, cmd.Flags().Lookup("debug"))
	assert.NotNil(t, cmd.Flags().Lookup("log-level"))
	assert.NotNil(t, cmd.Flags().Lookup("shutdown-timeout"))
	assert.NotNil(t, cmd.Flags().Lookup("logfile"))
}

func TestBindFlags(t *testing.T) {
	cmd := &cobra.Command{}
	DefineFlags(cmd)
	err := BindFlags(cmd)
	assert.NoError(t, err)
}

func TestBindFlags_Error(t *testing.T) {
	cmd := &cobra.Command{}
	err := BindFlags(cmd)
	assert.Error(t, err)
}
