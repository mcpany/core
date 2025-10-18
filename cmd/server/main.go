/*
 * Copyright 2025 Author(s) of MCP-XY
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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcpxy/core/pkg/app"
	"github.com/mcpxy/core/pkg/appconsts"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appRunner app.Runner = app.NewApplication()

// newRootCmd creates and configures the main command for the application. It
// sets up the command-line flags, environment variable binding, and the main
// execution logic for the server. It also adds subcommands for version and
// health checks.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   appconsts.Name,
		Short: "MCP-XY is a versatile proxy for backend services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonrpcPort := viper.GetString("jsonrpc-port")
			registrationPort := viper.GetString("grpc-port")
			stdio := viper.GetBool("stdio")
			configPaths := viper.GetStringSlice("config-paths")

			logLevel := slog.LevelInfo
			if viper.GetBool("debug") {
				logLevel = slog.LevelDebug
			}
			logging.Init(logLevel, os.Stdout)
			log := logging.GetLogger().With("service", "mcpxy")

			log.Info("Configuration", "jsonrpc-port", jsonrpcPort, "registration-port", registrationPort, "stdio", stdio, "config-paths", configPaths)
			if len(configPaths) > 0 {
				log.Info("Attempting to load services from config paths", "paths", configPaths)
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			osFs := afero.NewOsFs()

			shutdownTimeout := viper.GetDuration("shutdown-timeout")

			if err := appRunner.Run(ctx, osFs, stdio, jsonrpcPort, registrationPort, configPaths, shutdownTimeout); err != nil {
				log.Error("Application failed", "error", err)
				return err
			}
			log.Info("Shutdown complete.")
			return nil
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", appconsts.Name, appconsts.Version)
			if err != nil {
				return fmt.Errorf("failed to print version: %w", err)
			}
			return nil
		},
	}
	rootCmd.AddCommand(versionCmd)

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Run a health check against a running server",
		RunE: func(cmd *cobra.Command, args []string) error {
			port := viper.GetString("jsonrpc-port")
			return app.HealthCheck(port)
		},
	}
	rootCmd.AddCommand(healthCmd)

	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPXY")
	})

	rootCmd.PersistentFlags().String("jsonrpc-port", "50050", "Port for the JSON-RPC and HTTP registration server. Env: MCPXY_JSONRPC_PORT")
	if err := viper.BindPFlag("jsonrpc-port", rootCmd.PersistentFlags().Lookup("jsonrpc-port")); err != nil {
		fmt.Printf("Error binding jsonrpc-port flag: %v\n", err)
	}

	rootCmd.Flags().String("grpc-port", "", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPXY_GRPC_PORT")
	rootCmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPXY_STDIO")
	rootCmd.Flags().StringSlice("config-paths", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPXY_CONFIG_PATHS")
	rootCmd.Flags().Bool("debug", false, "Enable debug logging. Env: MCPXY_DEBUG")
	rootCmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPXY_SHUTDOWN_TIMEOUT")

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		fmt.Printf("Error binding command line flags: %v\n", err)
	}

	return rootCmd
}

// main is the entry point for the MCP-XY server application. It initializes and
// executes the root command, which handles command-line argument parsing,
// configuration, and the startup of the server.
func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
