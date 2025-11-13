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
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	viper.Set("jsonrpc-port", "8080")
	viper.Set("grpc-port", "9090")
	viper.Set("stdio", true)
	viper.Set("config-path", []string{"/etc/mcpany", "/opt/mcpany"})
	viper.Set("debug", true)
	viper.Set("logfile", "/var/log/mcpany.log")
	viper.Set("shutdown-timeout", 10*time.Second)

	cfg := Load()

	assert.Equal(t, "8080", cfg.JSONRPCPort)
	assert.Equal(t, "9090", cfg.RegistrationPort)
	assert.True(t, cfg.Stdio)
	assert.Equal(t, []string{"/etc/mcpany", "/opt/mcpany"}, cfg.ConfigPaths)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "/var/log/mcpany.log", cfg.LogFile)
	assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
}
