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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mcpany/core/pkg/logging"
	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/viper"
)

// Settings defines the global configuration for the application.
type Settings struct {
	v *viper.Viper
}

var (
	globalSettings *Settings
	once           sync.Once
)

// GlobalSettings returns the singleton instance of the global settings.
func GlobalSettings() *Settings {
	once.Do(func() {
		globalSettings = &Settings{}
	})
	return globalSettings
}

// Load initializes the global settings from a viper instance.
func (s *Settings) Load(v *viper.Viper) error {
	s.v = v
	return nil
}

// GRPCPort returns the gRPC port.
func (s *Settings) GRPCPort() string {
	return s.v.GetString("grpc-port")
}

// MCPListenAddress returns the MCP listen address.
func (s *Settings) MCPListenAddress() string {
	addr := s.v.GetString("mcp-listen-address")
	if !strings.Contains(addr, ":") {
		addr = "localhost:" + addr
	}
	return addr
}

// Stdio returns whether stdio mode is enabled.
func (s *Settings) Stdio() bool {
	return s.v.GetBool("stdio")
}

// ConfigPaths returns the paths to the configuration files.
func (s *Settings) ConfigPaths() []string {
	return s.v.GetStringSlice("config-path")
}

// IsDebug returns whether debug mode is enabled.
func (s *Settings) IsDebug() bool {
	return s.v.GetBool("debug")
}

// LogFile returns the path to the log file.
func (s *Settings) LogFile() string {
	return s.v.GetString("logfile")
}

// ShutdownTimeout returns the graceful shutdown timeout.
func (s *Settings) ShutdownTimeout() time.Duration {
	return s.v.GetDuration("shutdown-timeout")
}

// LogLevel returns the log level for the server.
func (s *Settings) LogLevel() v1.GlobalSettings_LogLevel {
	if s.IsDebug() {
		return v1.GlobalSettings_LOG_LEVEL_DEBUG
	}
	logLevel := s.v.GetString("log-level")
	switch strings.ToLower(logLevel) {
	case "debug":
		return v1.GlobalSettings_LOG_LEVEL_DEBUG
	case "info":
		return v1.GlobalSettings_LOG_LEVEL_INFO
	case "warn":
		return v1.GlobalSettings_LOG_LEVEL_WARN
	case "error":
		return v1.GlobalSettings_LOG_LEVEL_ERROR
	default:
		if logLevel != "" {
			logging.GetLogger().Warn(
				fmt.Sprintf(
					"Invalid log level specified: '%s'. Defaulting to INFO.",
					logLevel,
				),
			)
		}
		return v1.GlobalSettings_LOG_LEVEL_INFO
	}
}
