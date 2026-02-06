// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BindRootFlags binds the global and persistent command-line flags to the Viper configuration registry.
//
// Summary: Binds the global and persistent command-line flags to the Viper configuration registry.
//
// Parameters:
//   - cmd: *cobra.Command. The command instance to which the persistent flags will be attached.
//
// Returns:
//   None.
//
// Throws/Errors:
//   Exits the application with status code 1 if a flag binding operation fails
//   (e.g., if a flag with the same name already exists).
func BindRootFlags(cmd *cobra.Command) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MCPANY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server's bind address. Env: MCPANY_MCP_LISTEN_ADDRESS")
	cmd.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPANY_CONFIG_PATH")
	cmd.PersistentFlags().String("metrics-listen-address", "", "Address to expose Prometheus metrics on. If not specified, metrics are disabled. Env: MCPANY_METRICS_LISTEN_ADDRESS")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug logging. Env: MCPANY_DEBUG")
	cmd.PersistentFlags().String("log-level", "info", "Set the log level (debug, info, warn, error). Env: MCPANY_LOG_LEVEL")
	cmd.PersistentFlags().String("log-format", "text", "Set the log format (text, json). Env: MCPANY_LOG_FORMAT")
	cmd.PersistentFlags().String("logfile", "", "Path to a file to write logs to. If not set, logs are written to stdout.")
	cmd.PersistentFlags().StringSlice("set", []string{}, "Set configuration values (key=value). Supports nested keys with dots (e.g., upstream_services[0].http_service.address=http://localhost:8080)")

	if err := viper.BindPFlag("mcp-listen-address", cmd.PersistentFlags().Lookup("mcp-listen-address")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding mcp-listen-address flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("config-path", cmd.PersistentFlags().Lookup("config-path")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding config-path flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("metrics-listen-address", cmd.PersistentFlags().Lookup("metrics-listen-address")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding metrics-listen-address flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding debug flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("log-level", cmd.PersistentFlags().Lookup("log-level")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding log-level flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("log-format", cmd.PersistentFlags().Lookup("log-format")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding log-format flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("logfile", cmd.PersistentFlags().Lookup("logfile")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding logfile flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("set", cmd.PersistentFlags().Lookup("set")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding set flag: %v\n", err)
		os.Exit(1)
	}
}

// BindServerFlags binds server-specific command-line flags to the Viper configuration registry.
//
// Summary: Defines and binds flags specific to the server operation, such as port configurations and authentication keys.
//
// Parameters:
//   - cmd: *cobra.Command. The command instance to which the server flags will be attached.
//
// Returns:
//   None.
//
// Throws/Errors:
//   Exits the application with status code 1 if a flag binding operation fails.
func BindServerFlags(cmd *cobra.Command) {
	cmd.Flags().String("grpc-port", "", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPANY_GRPC_PORT")
	cmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPANY_STDIO")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("api-key", "", "API key for securing the MCP server. If set, all requests must include this key in the 'X-API-Key' header. Env: MCPANY_API_KEY")
	cmd.Flags().StringSlice("profiles", []string{"default"}, "Comma-separated list of active profiles. Env: MCPANY_PROFILES")
	cmd.Flags().String("db-path", "data/mcpany.db", "Path to the SQLite database file. Env: MCPANY_DB_PATH")

	if err := viper.BindPFlag("grpc-port", cmd.Flags().Lookup("grpc-port")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding grpc-port flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("stdio", cmd.Flags().Lookup("stdio")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding stdio flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("shutdown-timeout", cmd.Flags().Lookup("shutdown-timeout")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding shutdown-timeout flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("api-key", cmd.Flags().Lookup("api-key")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding api-key flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("profiles", cmd.Flags().Lookup("profiles")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding profiles flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("db-path", cmd.Flags().Lookup("db-path")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding db-path flag: %v\n", err)
		os.Exit(1)
	}
}

// BindFlags binds both root and server-specific command line flags to the Viper configuration registry.
//
// Summary: Orchestrates the binding of all necessary flags by delegating to BindRootFlags and BindServerFlags.
//
// Parameters:
//   - cmd: *cobra.Command. The command instance to which the flags will be attached.
//
// Returns:
//   None.
//
// Throws/Errors:
//   Exits the application with status code 1 if a flag binding operation fails.
func BindFlags(cmd *cobra.Command) {
	BindRootFlags(cmd)
	BindServerFlags(cmd)
}
