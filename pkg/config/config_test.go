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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBindFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindFlags(cmd)

	// Check a fiew flags to ensure they are bound
	assert.NotNil(t, cmd.PersistentFlags().Lookup("mcp-listen-address"))
	assert.NotNil(t, cmd.Flags().Lookup("grpc-port"))
	assert.NotNil(t, cmd.Flags().Lookup("stdio"))

	// Check that the values are correctly bound to viper
	cmd.Flags().Set("grpc-port", "8081")
	assert.Equal(t, "8081", viper.GetString("grpc-port"))

	cmd.Flags().Set("stdio", "true")
	assert.True(t, viper.GetBool("stdio"))
}
