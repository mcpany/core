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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BindRootFlags binds the global/persistent flags to viper.
func BindRootFlags(cmd *cobra.Command) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MCPANY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server's bind address. Env: MCPANY_MCP_LISTEN_ADDRESS")
	cmd.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH")
	cmd.PersistentFlags().String("metrics-listen-address", "", "Address to expose Prometheus metrics on. If not specified, metrics are disabled. Env: MCPANY_METRICS_LISTEN_ADDRESS")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug logging. Env: MCPANY_DEBUG")
	cmd.PersistentFlags().String("log-level", "info", "Set the log level (debug, info, warn, error). Env: MCPANY_LOG_LEVEL")
	cmd.PersistentFlags().String("logfile", "", "Path to a file to write logs to. If not set, logs are written to stdout.")

	if err := viper.BindPFlag("mcp-listen-address", cmd.PersistentFlags().Lookup("mcp-listen-address")); err != nil {
		fmt.Printf("Error binding mcp-listen-address flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("config-path", cmd.PersistentFlags().Lookup("config-path")); err != nil {
		fmt.Printf("Error binding config-path flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("metrics-listen-address", cmd.PersistentFlags().Lookup("metrics-listen-address")); err != nil {
		fmt.Printf("Error binding metrics-listen-address flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		fmt.Printf("Error binding debug flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("log-level", cmd.PersistentFlags().Lookup("log-level")); err != nil {
		fmt.Printf("Error binding log-level flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("logfile", cmd.PersistentFlags().Lookup("logfile")); err != nil {
		fmt.Printf("Error binding logfile flag: %v\n", err)
		os.Exit(1)
	}
}

// BindServerFlags binds server-specific flags to viper.
func BindServerFlags(cmd *cobra.Command) {
	cmd.Flags().String("grpc-port", "", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPANY_GRPC_PORT")
	cmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPANY_STDIO")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("api-key", "", "API key for securing the MCP server. If set, all requests must include this key in the 'X-API-Key' header. Env: MCPANY_API_KEY")

	if err := viper.BindPFlag("grpc-port", cmd.Flags().Lookup("grpc-port")); err != nil {
		fmt.Printf("Error binding grpc-port flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("stdio", cmd.Flags().Lookup("stdio")); err != nil {
		fmt.Printf("Error binding stdio flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("shutdown-timeout", cmd.Flags().Lookup("shutdown-timeout")); err != nil {
		fmt.Printf("Error binding shutdown-timeout flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("api-key", cmd.Flags().Lookup("api-key")); err != nil {
		fmt.Printf("Error binding api-key flag: %v\n", err)
		os.Exit(1)
	}
}

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	BindRootFlags(cmd)
	BindServerFlags(cmd)
}
