// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main is the mcpany server main
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/update"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	// Version is set at build time.
	Version              = "dev"
	appRunner app.Runner = app.NewApplication()
)

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
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the MCP Any server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				return err
			}

			if err := metrics.Initialize(); err != nil {
				logging.GetLogger().Error("Failed to initialize metrics", "error", err)
				os.Exit(1)
			}
			log := logging.GetLogger().With("service", "mcpany")

			bindAddress := cfg.MCPListenAddress()
			grpcPort := cfg.GRPCPort()
			stdio := cfg.Stdio()
			configPaths := cfg.ConfigPaths()

			log.Info("Configuration", "mcp-listen-address", bindAddress, "registration-port", grpcPort, "stdio", stdio, "config-path", configPaths)
			if len(configPaths) > 0 {
				log.Info("Attempting to load services from config path", "paths", configPaths)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				// Wait for an interrupt signal
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				<-sigChan
				log.Info("Received interrupt signal, shutting down...")
				cancel()
			}()

			// Start file watcher
			if !cfg.Stdio() {
				watcher, err := config.NewWatcher()
				if err != nil {
					return fmt.Errorf("failed to create file watcher: %w", err)
				}
				defer watcher.Close()

				go func() {
					if err := watcher.Watch(configPaths, func() {
						if err := appRunner.ReloadConfig(osFs, configPaths); err != nil {
							log.Error("Failed to reload config", "error", err)
						}
					}); err != nil {
						log.Error("Watcher failed", "error", err)
					}
				}()
			}

			shutdownTimeout := cfg.ShutdownTimeout()

			if metricsListenAddress := cfg.MetricsListenAddress(); metricsListenAddress != "" {
				go func() {
					log.Info("Starting metrics server", "address", metricsListenAddress)
					if err := metrics.StartServer(metricsListenAddress); err != nil {
						log.Error("Metrics server failed", "error", err)
					}
				}()
			}

			if err := appRunner.Run(ctx, osFs, stdio, bindAddress, grpcPort, configPaths, cfg.APIKey(), shutdownTimeout); err != nil {
				log.Error("Application failed", "error", err)
				return err
			}
			log.Info("Shutdown complete.")
			return nil
		},
	}
	config.BindServerFlags(runCmd)
	rootCmd.AddCommand(runCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpany",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", appconsts.Name, Version)
			if err != nil {
				return fmt.Errorf("failed to print version: %w", err)
			}
			return nil
		},
	}
	rootCmd.AddCommand(versionCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the application to the latest version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token := os.Getenv("GITHUB_TOKEN")
			var tc *http.Client
			if token != "" {
				ts := oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: token},
				)
				tc = oauth2.NewClient(context.Background(), ts)
			}
			updater := update.NewUpdater(tc)
			release, available, err := updater.CheckForUpdate(context.Background(), "mcpany", "core", Version)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !available {
				fmt.Println("You are already running the latest version.")
				return nil
			}

			fmt.Printf("A new version is available: %s. Updating...\n", release.GetTagName())

			assetName := fmt.Sprintf("server-%s-%s", runtime.GOOS, runtime.GOARCH)
			checksumsAssetName := "checksums.txt"

			fs := afero.NewOsFs()
			executablePath, _ := cmd.Flags().GetString("path")
			if executablePath == "" {
				var err error
				executablePath, err = os.Executable()
				if err != nil {
					return fmt.Errorf("failed to get executable path: %w", err)
				}
			}

			if err := updater.UpdateTo(context.Background(), fs, executablePath, release, assetName, checksumsAssetName); err != nil {
				return fmt.Errorf("failed to update: %w", err)
			}

			fmt.Println("Update successful.")
			return nil
		},
	}
	updateCmd.Flags().String("path", "", "Path to the binary to update. Defaults to the current executable.")
	rootCmd.AddCommand(updateCmd)

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Run a health check against a running server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, fs); err != nil {
				return err
			}
			addr := cfg.MCPListenAddress()

			if !strings.Contains(addr, ":") {
				addr = "localhost:" + addr
			}
			timeout, _ := cmd.Flags().GetDuration("timeout")
			return app.HealthCheck(cmd.OutOrStdout(), addr, timeout)
		},
	}
	healthCmd.Flags().Duration("timeout", 5*time.Second, "Timeout for the health check.")
	rootCmd.AddCommand(healthCmd)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("MCP Any CLI: Configuration Generator")

			generator := config.NewGenerator()
			configData, err := generator.Generate()
			if err != nil {
				return err
			}

			fmt.Println("\nGenerated configuration:")
			fmt.Print(string(configData))

			return nil
		},
	}
	configCmd.AddCommand(generateCmd)

	docCmd := &cobra.Command{
		Use:   "doc",
		Short: "Generate Markdown documentation for the configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfgSettings := config.GlobalSettings()
			if err := cfgSettings.Load(cmd, osFs); err != nil {
				return err
			}

			store := config.NewFileStore(osFs, cfgSettings.ConfigPaths())
			cfg, err := config.LoadServices(context.Background(), store, "server")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			doc, err := config.GenerateDocumentation(context.Background(), cfg)
			if err != nil {
				return err
			}

			fmt.Println(doc)
			return nil
		},
	}
	configCmd.AddCommand(docCmd)

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			configs, err := config.LoadResolvedConfig(context.Background(), store)
			if err != nil {
				return fmt.Errorf("failed to load configurations for validation: %w", err)
			}

			var allErrors []string
			if validationErrors := config.Validate(context.Background(), configs, config.Server); len(validationErrors) > 0 {
				for _, e := range validationErrors {
					allErrors = append(allErrors, e.Error())
				}
			}

			if len(allErrors) > 0 {
				return fmt.Errorf("configuration validation failed with errors: \n- %s", strings.Join(allErrors, "\n- "))
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid.")
			if err != nil {
				return fmt.Errorf("failed to print validation success message: %w", err)
			}
			return nil
		},
	}
	configCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(configCmd)

	config.BindRootFlags(rootCmd)

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
