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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DefineFlags defines the command line flags.
func DefineFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server's bind address. Env: MCPANY_MCP_LISTEN_ADDRESS")
	cmd.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH")
	cmd.Flags().String("grpc-port", "", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPANY_GRPC_PORT")
	cmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPANY_STDIO")
	cmd.Flags().Bool("debug", false, "Enable debug logging. Env: MCPANY_DEBUG")
	cmd.Flags().String("log-level", "info", "Set the log level (debug, info, warn, error). Env: MCPANY_LOG_LEVEL")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("logfile", "", "Path to a file to write logs to. If not set, logs are written to stdout.")
}

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) error {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
	})

	if err := viper.BindPFlag("mcp-listen-address", cmd.PersistentFlags().Lookup("mcp-listen-address")); err != nil {
		return fmt.Errorf("Error binding mcp-listen-address flag: %w", err)
	}
	if err := viper.BindPFlag("config-path", cmd.PersistentFlags().Lookup("config-path")); err != nil {
		return fmt.Errorf("Error binding config-path flag: %w", err)
	}

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("Error binding command line flags: %w", err)
	}
	return nil
}
