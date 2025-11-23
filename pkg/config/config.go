// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
	})

	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server's bind address. Env: MCPANY_MCP_LISTEN_ADDRESS")
	cmd.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH")
	if err := viper.BindPFlag("mcp-listen-address", cmd.PersistentFlags().Lookup("mcp-listen-address")); err != nil {
		fmt.Printf("Error binding mcp-listen-address flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("config-path", cmd.PersistentFlags().Lookup("config-path")); err != nil {
		fmt.Printf("Error binding config-path flag: %v\n", err)
		os.Exit(1)
	}

	cmd.Flags().String("grpc-port", "", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPANY_GRPC_PORT")
	cmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPANY_STDIO")
	cmd.Flags().Bool("debug", false, "Enable debug logging. Env: MCPANY_DEBUG")
	cmd.Flags().String("log-level", "info", "Set the log level (debug, info, warn, error). Env: MCPANY_LOG_LEVEL")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("logfile", "", "Path to a file to write logs to. If not set, logs are written to stdout.")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		fmt.Printf("Error binding command line flags: %v\n", err)
		os.Exit(1)
	}
}
