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
	"time"

	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
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
	cmd.Flags().Bool("debug", false, "Enable debug logging. Env: MCPANY_DEBUG")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("logfile", "", "Path to a file to write logs to. If not set, logs are written to stdout.")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		fmt.Printf("Error binding command line flags: %v\n", err)
		os.Exit(1)
	}
}

// GetBindAddress returns the bind address for the server.
// It prioritizes the config file over the command line flag if the flag is not explicitly set.
func GetBindAddress(cmd *cobra.Command, fs afero.Fs) (string, error) {
	bindAddress := viper.GetString("jsonrpc-port")
	configPaths := viper.GetStringSlice("config-path")

	if !cmd.Flags().Changed("jsonrpc-port") && len(configPaths) > 0 {
		store := NewFileStore(fs, configPaths)
		cfg, err := LoadServices(store, "server")
		if err != nil {
			return "", fmt.Errorf("failed to load services from config: %w", err)
		}
		if cfg.GetGlobalSettings().GetBindAddress() != "" {
			bindAddress = cfg.GetGlobalSettings().GetBindAddress()
		}
	}
	return bindAddress, nil
}

// GetGRPCPort returns the gRPC port.
func GetGRPCPort() string {
	return viper.GetString("grpc-port")
}

// GetStdio returns whether stdio mode is enabled.
func GetStdio() bool {
	return viper.GetBool("stdio")
}

// GetConfigPaths returns the paths to the configuration files.
func GetConfigPaths() []string {
	return viper.GetStringSlice("config-path")
}

// IsDebug returns whether debug mode is enabled.
func IsDebug() bool {
	return viper.GetBool("debug")
}

// GetLogFile returns the path to the log file.
func GetLogFile() string {
	return viper.GetString("logfile")
}

// GetShutdownTimeout returns the graceful shutdown timeout.
func GetShutdownTimeout() time.Duration {
	return viper.GetDuration("shutdown-timeout")
}

// GetLogLevel returns the log level for the server.
func GetLogLevel() v1.GlobalSettings_LogLevel {
	if IsDebug() {
		return v1.GlobalSettings_DEBUG
	}
	return v1.GlobalSettings_INFO
}
