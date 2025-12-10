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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Settings defines the global configuration for the application.
type Settings struct {
	proto           *v1.GlobalSettings
	grpcPort        string
	stdio           bool
	configPaths     []string
	debug           bool
	logLevel        string
	logFile         string
	shutdownTimeout time.Duration
	fs              afero.Fs
	cmd             *cobra.Command
}

var (
	globalSettings *Settings
	once           sync.Once
)

// GlobalSettings returns the singleton instance of the global settings.
func GlobalSettings() *Settings {
	once.Do(func() {
		globalSettings = &Settings{
			proto: &v1.GlobalSettings{},
		}
	})
	return globalSettings
}

// Load initializes the global settings from the command line and config files.
func (s *Settings) Load(cmd *cobra.Command, fs afero.Fs) error {
	s.cmd = cmd
	s.fs = fs

	s.grpcPort = viper.GetString("grpc-port")
	s.stdio = viper.GetBool("stdio")
	s.configPaths = viper.GetStringSlice("config-path")
	s.debug = viper.GetBool("debug")
	s.logLevel = viper.GetString("log-level")
	s.logFile = viper.GetString("logfile")
	s.shutdownTimeout = viper.GetDuration("shutdown-timeout")

	// Special handling for MCPListenAddress to respect config file precedence
	mcpListenAddress := viper.GetString("mcp-listen-address")
	apiKey := viper.GetString("api-key")

	if len(s.configPaths) > 0 {
		store := NewFileStore(fs, s.configPaths)
		cfg, err := LoadServices(store, "server")
		if err != nil {
			return fmt.Errorf("failed to load services from config: %w", err)
		}
		if !cmd.Flags().Changed("mcp-listen-address") && cfg.GetGlobalSettings().GetMcpListenAddress() != "" {
			mcpListenAddress = cfg.GetGlobalSettings().GetMcpListenAddress()
		}
		if !cmd.Flags().Changed("api-key") && cfg.GetGlobalSettings().GetApiKey() != "" {
			apiKey = cfg.GetGlobalSettings().GetApiKey()
		}
	}

	if apiKey != "" && len(apiKey) < 16 {
		return fmt.Errorf("API key must be at least 16 characters long")
	}

	s.proto.SetMcpListenAddress(mcpListenAddress)
	s.proto.SetLogLevel(s.LogLevel())
	s.proto.SetApiKey(apiKey)

	return nil
}

// GRPCPort returns the gRPC port.
func (s *Settings) GRPCPort() string {
	return s.grpcPort
}

// MCPListenAddress returns the MCP listen address.
func (s *Settings) MCPListenAddress() string {
	addr := s.proto.GetMcpListenAddress()
	if !strings.Contains(addr, ":") {
		addr = "localhost:" + addr
	}
	return addr
}

// MetricsListenAddress returns the metrics listen address.
func (s *Settings) MetricsListenAddress() string {
	return viper.GetString("metrics-listen-address")
}

// Stdio returns whether stdio mode is enabled.
func (s *Settings) Stdio() bool {
	return s.stdio
}

// ConfigPaths returns the paths to the configuration files.
func (s *Settings) ConfigPaths() []string {
	return s.configPaths
}

// IsDebug returns whether debug mode is enabled.
func (s *Settings) IsDebug() bool {
	return s.debug
}

// LogFile returns the path to the log file.
func (s *Settings) LogFile() string {
	return s.logFile
}

// ShutdownTimeout returns the graceful shutdown timeout.
func (s *Settings) ShutdownTimeout() time.Duration {
	return s.shutdownTimeout
}

// APIKey returns the API key for the server.
func (s *Settings) APIKey() string {
	if s.proto.GetApiKey() != "" {
		return s.proto.GetApiKey()
	}
	return viper.GetString("api-key")
}

func (s *Settings) LogLevel() v1.GlobalSettings_LogLevel {
	if s.IsDebug() {
		return v1.GlobalSettings_LOG_LEVEL_DEBUG
	}
	switch strings.ToLower(s.logLevel) {
	case "debug":
		return v1.GlobalSettings_LOG_LEVEL_DEBUG
	case "info":
		return v1.GlobalSettings_LOG_LEVEL_INFO
	case "warn":
		return v1.GlobalSettings_LOG_LEVEL_WARN
	case "error":
		return v1.GlobalSettings_LOG_LEVEL_ERROR
	default:
		if s.logLevel != "" {
			logging.GetLogger().Warn(
				fmt.Sprintf(
					"Invalid log level specified: '%s'. Defaulting to INFO.",
					s.logLevel,
				),
			)
		}
		return v1.GlobalSettings_LOG_LEVEL_INFO
	}
}
