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

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/appconsts"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var appRunner app.Runner = app.NewApplication()

// newRootCmd creates and configures the main command for the application.
// It sets up the command-line flags for configuring the server, such as ports,
// configuration paths, and operational modes (e.g., stdio). It also handles
// environment variable binding, allowing for configuration through both flags
// and environment variables.
//
// The command's main execution logic initializes the logger, sets up a context
// for graceful shutdown, and starts the application runner with the parsed
// configuration.
//
// Additionally, it adds subcommands for `version` to print the application's
// version and `health` to perform a health check against a running server.
//
// Returns the configured root command, ready to be executed.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   appconsts.Name,
		Short: "MCP Any is a versatile proxy for backend services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.LoadServerConfig()

			logLevel := slog.LevelInfo
			if cfg.Debug {
				logLevel = slog.LevelDebug
			}

			var logOutput io.Writer = os.Stdout
			if cfg.LogFile != "" {
				f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to open logfile: %w", err)
				}
				defer f.Close()
				logOutput = f
			} else if cfg.Stdio {
				logOutput = io.Discard // Disable logging in stdio mode to keep the channel clean for JSON-RPC
			}
			logging.Init(logLevel, logOutput)

			metrics.Initialize()
			log := logging.GetLogger().With("service", "mcpany")

			log.Info("Configuration", "jsonrpc-port", cfg.JSONRPCPort, "registration-port", cfg.GRPCPort, "stdio", cfg.Stdio, "config-path", cfg.ConfigPaths)
			if len(cfg.ConfigPaths) > 0 {
				log.Info("Attempting to load services from config path", "paths", cfg.ConfigPaths)
			}

			osFs := afero.NewOsFs()
			bindAddress := cfg.JSONRPCPort

			// If the jsonrpc-port flag is not explicitly set, we'll check the config file.
			if !cmd.Flags().Changed("jsonrpc-port") && len(cfg.ConfigPaths) > 0 {
				store := config.NewFileStore(osFs, cfg.ConfigPaths)
				loadedCfg, err := config.LoadServices(store, "server")
				if err != nil {
					return fmt.Errorf("failed to load services from config: %w", err)
				}
				if loadedCfg.GetGlobalSettings().GetBindAddress() != "" {
					bindAddress = loadedCfg.GetGlobalSettings().GetBindAddress()
				}
			}

			if !strings.Contains(bindAddress, ":") {
				bindAddress = "localhost:" + bindAddress
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			if err := appRunner.Run(ctx, osFs, cfg.Stdio, bindAddress, cfg.GRPCPort, cfg.ConfigPaths, cfg.ShutdownTimeout); err != nil {
				log.Error("Application failed", "error", err)
				return err
			}
			log.Info("Shutdown complete.")
			return nil
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpany",
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
			cfg := config.LoadServerConfig()
			fs := afero.NewOsFs()
			addr := cfg.JSONRPCPort
			if !cmd.Flags().Changed("jsonrpc-port") && len(cfg.ConfigPaths) > 0 {
				store := config.NewFileStore(fs, cfg.ConfigPaths)
				loadedCfg, err := config.LoadServices(store, "server")
				if err != nil {
					return fmt.Errorf("failed to load services from config: %w", err)
				}
				if loadedCfg.GetGlobalSettings().GetBindAddress() != "" {
					addr = loadedCfg.GetGlobalSettings().GetBindAddress()
				}
			}

			if !strings.Contains(addr, ":") {
				addr = "localhost:" + addr
			}
			timeout, _ := cmd.Flags().GetDuration("timeout")
			return app.HealthCheck(cmd.OutOrStdout(), addr, timeout)
		},
	}
	healthCmd.Flags().Duration("timeout", 5*time.Second, "Timeout for the health check.")
	rootCmd.AddCommand(healthCmd)

	config.BindFlags(rootCmd)

	return rootCmd
}

// main is the entry point for the MCP Any server application. It initializes and
// executes the root command, which is responsible for parsing command-line
// arguments, loading configuration, and starting the server.
//
// The application will exit with a non-zero status code if an error occurs
// during the execution of the command.
func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
