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
	"os"
	"strings"
	"sync"
	"time"

	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	globalFs   = afero.NewOsFs()
	configOnce sync.Once
	loadErr    error
)

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_")) // Allow log-file to be set by MCPANY_LOG_FILE
	})

	cmd.PersistentFlags().String("jsonrpc-port", "50050", "Port for the JSON-RPC and HTTP registration server. Env: MCPANY_JSONRPC_PORT")
	cmd.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH")
	if err := viper.BindPFlag("jsonrpc-port", cmd.PersistentFlags().Lookup("jsonrpc-port")); err != nil {
		fmt.Printf("Error binding jsonrpc-port flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("config-path", cmd.PersistentFlags().Lookup("config-path")); err != nil {
		fmt.Printf("Error binding config-path flag: %v\n", err)
		os.Exit(1)
	}

	cmd.Flags().String("grpc-port", "", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPANY_GRPC_PORT")
	cmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPANY_STDIO")
	cmd.Flags().Bool("debug", false, "Enable debug logging (equivalent to --log-level=DEBUG). Env: MCPANY_DEBUG")
	cmd.Flags().String("log-level", "INFO", "Set the log level (e.g., DEBUG, INFO, WARN, ERROR). Env: MCPANY_LOG_LEVEL")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("log-file", "", "Path to a file to write logs to. If not set, logs are written to stdout. Env: MCPANY_LOG_FILE")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		fmt.Printf("Error binding command line flags: %v\n", err)
		os.Exit(1)
	}
}

// mergeConfig loads the configuration from the file and sets the values
// as defaults in Viper. This allows Viper's precedence (flags > env > defaults)
// to work correctly with the config file values.
func mergeConfig() {
	configOnce.Do(func() {
		configPaths := viper.GetStringSlice("config-path")
		if len(configPaths) == 0 {
			return // No config path, so nothing to merge.
		}

		store := NewFileStore(globalFs, configPaths)
		cfg, err := LoadServices(store, "server")
		if err != nil {
			// If the file doesn't exist or is invalid, we'll just use viper's defaults/flags.
			// We store the error in case any caller wants to inspect it.
			loadErr = err
			return
		}

		if cfg != nil && cfg.GetGlobalSettings() != nil {
			settings := cfg.GetGlobalSettings()
			if settings.GetJsonrpcPort() != "" {
				viper.SetDefault("jsonrpc-port", settings.GetJsonrpcPort())
			}
			if settings.GetGrpcPort() != "" {
				viper.SetDefault("grpc-port", settings.GetGrpcPort())
			}
			viper.SetDefault("stdio", settings.GetStdio())
			if settings.GetLogFile() != "" {
				viper.SetDefault("log-file", settings.GetLogFile())
			}
			// Set log-level from config file. Note that the --debug flag will override this.
			viper.SetDefault("log-level", settings.GetLogLevel().String())
			if shutdownTimeout := settings.GetShutdownTimeout(); shutdownTimeout != nil {
				viper.SetDefault("shutdown-timeout", shutdownTimeout.AsDuration())
			}
		}
	})
}

// GetBindAddress returns the bind address for the server.
// It prioritizes flags > env vars > config file > defaults.
func GetBindAddress(cmd *cobra.Command, fs afero.Fs) (string, error) {
	// Allow overriding the global filesystem for testing purposes.
	globalFs = fs
	mergeConfig()
	return viper.GetString("jsonrpc-port"), loadErr
}

// GetGRPCPort returns the gRPC port.
func GetGRPCPort() string {
	mergeConfig()
	return viper.GetString("grpc-port")
}

// GetStdio returns whether stdio mode is enabled.
func GetStdio() bool {
	mergeConfig()
	return viper.GetBool("stdio")
}

// GetConfigPaths returns the paths to the configuration files.
func GetConfigPaths() []string {
	return viper.GetStringSlice("config-path")
}

// IsDebug returns whether debug mode is enabled.
func IsDebug() bool {
	return GetLogLevel() == v1.GlobalSettings_DEBUG
}

// GetLogFile returns the path to the log file.
func GetLogFile() string {
	mergeConfig()
	return viper.GetString("log-file")
}

// GetShutdownTimeout returns the graceful shutdown timeout.
func GetShutdownTimeout() time.Duration {
	mergeConfig()
	return viper.GetDuration("shutdown-timeout")
}

// GetLogLevel returns the log level for the server.
func GetLogLevel() v1.GlobalSettings_LogLevel {
	mergeConfig()
	// The debug flag takes highest precedence.
	if viper.GetBool("debug") {
		return v1.GlobalSettings_DEBUG
	}
	// Then, we check the log-level flag/env/config value.
	levelStr := strings.ToUpper(viper.GetString("log-level"))
	if level, ok := v1.GlobalSettings_LogLevel_value[levelStr]; ok {
		return v1.GlobalSettings_LogLevel(level)
	}
	// Default to INFO if the value is invalid.
	return v1.GlobalSettings_INFO
}
