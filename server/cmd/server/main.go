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

	"github.com/joho/godotenv"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/mcpany/core/server/pkg/lint"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/update"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

var (
	// Version is set at build time.
	Version              = "dev"
	appRunner app.Runner = app.NewApplication()
)

const (
	iconOk      = "✅"
	iconWarning = "⚠️ "
	iconError   = "❌"
	iconSkipped = "⏭️ "
)

// loadEnv loads environment variables from a .env file.
//
// Summary: Loads environment variables from a specified file or the default .env file.
//
// Parameters:
//   - cmd: *cobra.Command. The command instance to check for the --env-file flag.
//
// Returns:
//   - error: An error if the specified file cannot be loaded or if the default .env file is present but invalid.
//
// Side Effects:
//   - Modifies the process environment variables.
func loadEnv(cmd *cobra.Command) error {
	envFile, _ := cmd.Flags().GetString("env-file")
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return fmt.Errorf("failed to load env file %q: %w", envFile, err)
		}
		return nil
	}

	// Try loading default .env
	// We check for existence first to avoid silence on parse errors for an existing file
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("failed to parse .env file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		// Return error if stat failed for reason other than not exist (e.g. permission)
		return fmt.Errorf("failed to check for .env file: %w", err)
	}

	return nil
}

// newRootCmd creates and configures the main command for the application.
//
// Summary: Initializes the root Cobra command with all subcommands and flags.
//
// Returns:
//   - *cobra.Command: The configured root command.
//
// Side Effects:
//   - Configures global flags and command structure.
//   - Sets up persistent pre-run hooks for environment loading.
func newRootCmd() *cobra.Command { //nolint:gocyclo // Main entry point, expected to be complex
	rootCmd := &cobra.Command{
		Use:   appconsts.Name,
		Short: "MCP Any is a versatile proxy for backend services.",
		// We use PersistentPreRunE to load the .env file before any command runs.
		// Note: Cobra's OnInitialize is not used here, so this is safe.
		// If OnInitialize were used, we would need to ensure godotenv is loaded before Viper reads env vars.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return loadEnv(cmd)
		},
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().String("env-file", "", "Path to .env file to load environment variables from")

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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var logStoreCloser func() error
			defer func() {
				if logStoreCloser != nil {
					_ = logStoreCloser()
				}
			}()

			// Track 2: Product Gap - Log Persistence
			// Initialize SQLite Log Store if DB path is configured (default is data/mcpany.db)
			if cfg.DBPath() != "" {
				store, err := logging.NewSQLiteLogStore(cfg.DBPath())
				if err != nil {
					log.Warn("Failed to initialize SQLite log store", "error", err)
				} else {
					logging.SetStore(store)
					logStoreCloser = store.Close
					if err := logging.HydrateFromStore(store); err != nil {
						log.Warn("Failed to hydrate logs from store", "error", err)
					} else {
						log.Info("Hydrated log history from SQLite store")
					}
				}
			} else if cfg.LogFormat() == configv1.GlobalSettings_LOG_FORMAT_JSON && cfg.LogFile() != "" {
				// Fallback to file hydration if no DB but LogFile is present (legacy)
				go func() {
					if err := logging.HydrateFromFile(cfg.LogFile()); err != nil {
						log.Warn("Failed to hydrate logs from file", "error", err)
					} else {
						log.Info("Hydrated log history from file")
					}
				}()
			}

			// Track 1: Friction Fighter - Verify config files exist before proceeding
			if len(configPaths) > 0 {
				log.Info("Attempting to load services from config path", "paths", configPaths)
				for _, path := range configPaths {
					if strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://") {
						continue
					}
					if _, err := osFs.Stat(path); os.IsNotExist(err) {
						return fmt.Errorf("❌ Configuration file not found: %s\n\n💡 Tip: You can generate a default configuration using:\n   %s config generate > %s", path, appconsts.Name, path)
					} else if err != nil {
						return fmt.Errorf("❌ Failed to access configuration file %s: %w", path, err)
					}
				}
			} else {
				log.Warn("⚠️  No configuration files provided. Server will run but no tools will be available.")
			}

			go func() {
				// Wait for an interrupt signal or context cancellation
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				defer signal.Stop(sigChan)

				select {
				case <-sigChan:
					log.Info("Received interrupt signal, shutting down...")
					cancel()
				case <-ctx.Done():
					// Context cancelled, exit goroutine
				}
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
						if err := appRunner.ReloadConfig(ctx, osFs, configPaths); err != nil {
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

			if cfg.APIKey() == "" {
				log.Warn("⚠️  Running without a global API key. This is NOT recommended for production deployment.")
			}

			// Track 1: Friction Fighter - Strict Mode
			strict, _ := cmd.Flags().GetBool("strict")
			if strict {
				log.Info("Running in strict mode: validating configuration and upstream connectivity...")
				store := config.NewFileStore(osFs, configPaths)
				configs, err := config.LoadResolvedConfig(ctx, store)
				if err != nil {
					return fmt.Errorf("failed to load configuration for strict validation: %w", err)
				}

				results := doctor.RunChecks(ctx, configs)
				hasErrors := false
				for _, res := range results {
					var icon string
					switch res.Status {
					case doctor.StatusOk:
						icon = iconOk
						log.Info(fmt.Sprintf("%s [%s] %s: %s", icon, res.Status, res.ServiceName, res.Message))
					case doctor.StatusWarning:
						icon = iconWarning
						log.Warn(fmt.Sprintf("%s [%s] %s: %s", icon, res.Status, res.ServiceName, res.Message))
					case doctor.StatusError:
						icon = iconError
						log.Error(fmt.Sprintf("%s [%s] %s: %s", icon, res.Status, res.ServiceName, res.Message))
						hasErrors = true
					case doctor.StatusSkipped:
						icon = iconSkipped
						log.Info(fmt.Sprintf("%s [%s] %s: %s", icon, res.Status, res.ServiceName, res.Message))
					}
				}

				if hasErrors {
					return fmt.Errorf("strict mode validation failed: one or more upstream services are unreachable or misconfigured")
				}
				log.Info("Strict mode validation passed.")
			}

			// Track 1: Friction Fighter - Startup Banner
			// We defer the banner printing until the app is actually running to ensure ports are bound.
			// However, appRunner.Run blocks, so we need to hook into the startup process.
			// appRunner.Run takes a startupCallback indirectly via runServerMode logic inside app.
			// But the Runner interface doesn't expose it.
			// Instead, we will print "Starting..." here, and the banner logic should ideally handle the "Ready" state.
			// Since we can't easily modify appRunner.Run signature in this scope without larger refactor,
			// we will rely on log messages for now, or print a banner BEFORE Run saying "Initializing..."

			if !stdio {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n🚀 Starting %s v%s...\n", appconsts.Name, Version)
				if len(configPaths) > 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), "📋 Loading configuration from: %s\n", strings.Join(configPaths, ", "))
				}
			}

			if err := appRunner.Run(app.RunOptions{
				Ctx:             ctx,
				Fs:              osFs,
				Stdio:           stdio,
				JSONRPCPort:     bindAddress,
				GRPCPort:        grpcPort,
				ConfigPaths:     configPaths,
				APIKey:          cfg.APIKey(),
				ShutdownTimeout: shutdownTimeout,
			}); err != nil {
				log.Error("Application failed", "error", err)
				return err
			}
			log.Info("Shutdown complete.")
			return nil
		},
	}
	runCmd.Flags().Bool("strict", false, "Run in strict mode (validate upstream connectivity before starting)")
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
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			// We ignoring error here because even if config loading fails (e.g. missing file),
			// we might still want to proceed with defaults or env vars,
			// primarily for the GITHUB_TOKEN/GITHUB_API_URL.
			_ = cfg.Load(cmd, osFs)

			token := os.Getenv("GITHUB_TOKEN")
			var tc *http.Client
			if token != "" {
				ts := oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: token},
				)
				tc = oauth2.NewClient(context.Background(), ts)
			}

			githubURL := cfg.GithubAPIURL()
			if githubURL == "" {
				githubURL = os.Getenv("GITHUB_API_URL")
			}

			updater := update.NewUpdater(tc, githubURL)
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

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the health and connectivity of upstream services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration load failed: %w", err)
			}
			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			store.SetIgnoreMissingEnv(true)
			configs, err := config.LoadResolvedConfig(context.Background(), store)
			if err != nil {
				return fmt.Errorf("failed to load configurations: %w", err)
			}

			fmt.Println("Running doctor checks...")
			results := doctor.RunChecks(context.Background(), configs)

			doctor.PrintResults(cmd.OutOrStdout(), results)

			hasErrors := false
			for _, res := range results {
				if res.Status == doctor.StatusError {
					hasErrors = true
					break
				}
			}

			if hasErrors {
				return fmt.Errorf("doctor checks failed with errors")
			}
			fmt.Println("All checks passed!")
			return nil
		},
	}
	rootCmd.AddCommand(doctorCmd)

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

	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the JSON Schema for configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			schemaBytes, err := config.GenerateJSONSchemaBytes()
			if err != nil {
				return fmt.Errorf("failed to generate schema: %w", err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(schemaBytes))
			if err != nil {
				return fmt.Errorf("failed to print schema: %w", err)
			}
			return nil
		},
	}
	configCmd.AddCommand(schemaCmd)

	checkCmd := &cobra.Command{
		Use:   "check [file]",
		Short: "Check a configuration file against the JSON Schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			filename := args[0]
			data, err := os.ReadFile(filename) //nolint:gosec // Intended file inclusion
			if err != nil {
				return fmt.Errorf("failed to read file %q: %w", filename, err)
			}

			// Unmarshal as YAML (superset of JSON) into generic map
			var rawConfig map[string]interface{}
			if err := yaml.Unmarshal(data, &rawConfig); err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}

			if err := config.ValidateConfigAgainstSchema(rawConfig); err != nil {
				return fmt.Errorf("configuration schema validation failed: %w", err)
			}

			fmt.Println("Configuration schema is valid.")
			return nil
		},
	}
	configCmd.AddCommand(checkCmd)

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

			checkConnection, _ := cmd.Flags().GetBool("check-connection")
			if checkConnection {
				fmt.Println("Running connection checks...")
				results := doctor.RunChecks(context.Background(), configs)
				doctor.PrintResults(cmd.OutOrStdout(), results)

				hasDoctorErrors := false
				for _, res := range results {
					if res.Status == doctor.StatusError {
						hasDoctorErrors = true
						break
					}
				}
				if hasDoctorErrors {
					return fmt.Errorf("connection checks failed")
				}
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid.")
			if err != nil {
				return fmt.Errorf("failed to print validation success message: %w", err)
			}
			return nil
		},
	}
	validateCmd.Flags().Bool("check-connection", false, "Run connectivity checks for upstream services")
	configCmd.AddCommand(validateCmd)

	lintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint the configuration file for best practices and security issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration load failed: %w", err)
			}
			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			configs, err := config.LoadResolvedConfig(context.Background(), store)
			if err != nil {
				return fmt.Errorf("failed to load configurations for linting: %w", err)
			}

			linter := lint.NewLinter(configs)
			results, err := linter.Run(context.Background())
			if err != nil {
				return fmt.Errorf("linting failed: %w", err)
			}

			if len(results) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Lint passed! No issues found.")
				return nil
			}

			hasErrors := false
			for _, result := range results {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.String())
				if result.Severity == lint.Error {
					hasErrors = true
				}
			}

			if hasErrors {
				return fmt.Errorf("linting failed with errors")
			}
			return nil
		},
	}
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(configCmd)

	config.BindRootFlags(rootCmd)

	return rootCmd
}

// main is the entry point for the MCP Any server application.
//
// Summary: Executes the root command to start the application.
//
// Side Effects:
//   - Parsed command line arguments.
//   - Starts the HTTP/gRPC server.
//   - May exit the process with status code 1 on error.
func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
