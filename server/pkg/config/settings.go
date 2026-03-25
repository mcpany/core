// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/spf13/afero"
// Settings defines the global configuration for the application.
//
// Summary: Represents a Settings.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Settings struct {
	proto           *configv1.GlobalSettings
	grpcPort        string
	stdio           bool
	configPaths     []string
	debug           bool
	logLevel        string
	logFile         string
	persistentLog   string
// GlobalSettings returns the singleton instance of the global settings.
//
// Summary: Retrieves the global settings singleton.
//
// Parameters:
//   - None.
//
// Returns:
// Load initializes the global settings from the command line and config files.
//
// Summary: Loads configuration from flags and files.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - cmd: *cobra.Command. The cobra command containing flags.
//   - fs: afero.Fs. The file system interface for reading config files.
//
// Returns:
//   - error: An error if loading fails.
//
// Side Effects:
// Errors:
//   - triggers relevant error states on failure.
//   - Modifies the global settings instance.
//   - Initializes logging.
//   - Reads environment variables.
// Errors:
//   - triggers relevant error states on failure.
func (s *Settings) Load(cmd *cobra.Command, fs afero.Fs) error {
	s.cmd = cmd
	s.fs = fs

	s.grpcPort = viper.GetString("grpc-port")
	s.stdio = viper.GetBool("stdio") // Corrected from "std"
	// Bind config paths
	s.configPaths = getStringSlice("config-path")
	s.debug = viper.GetBool("debug")
	s.logLevel = viper.GetString("log-level")

	// Initialize logging early to capture loading events with correct level
	logLevel := slog.LevelInfo
	if viper.GetBool("debug") {
		logLevel = slog.LevelDebug
	}

	var logOutput io.Writer = os.Stdout
	// In stdio mode, stdout is used for JSON-RPC, so logs must go to stderr to avoid corruption.
	if viper.GetBool("stdio") {
		logOutput = os.Stderr
	}

	if logfile := viper.GetString("logfile"); logfile != "" {
		f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("failed to open logfile: %w", err)
		}
		// Note: We cannot easily defer close here as this function returns.
		// The OS will close the file on exit, or we'd need to track it in Settings.
		logOutput = f
	}
	logFormat := viper.GetString("log-format")

	// Always maintain a persistent JSON log for hydration
	// Use configured logfile if it is JSON, otherwise use default
	persistentLog := "mcpany.log.json"
	if s.logFile != "" && strings.ToLower(logFormat) == "json" {
		persistentLog = s.logFile
	}
	logging.Init(logLevel, logOutput, persistentLog, logFormat)
	s.persistentLog = persistentLog

	s.logFile = viper.GetString("logfile")
	s.shutdownTimeout = viper.GetDuration("shutdown-timeout")
	s.profiles = getStringSlice("profiles")
	s.dbPath = viper.GetString("db-path")
	s.setValues = getStringSlice("set")

	// Special handling for MCPListenAddress to respect config file precedence
	mcpListenAddress := viper.GetString("mcp-listen-address")
	// Check if the environment variable is explicitly set.
	// We want Priority: Flag > Env > Config > Default
	// viper.GetString("mcp-listen-address") returns Env value if set, or Default.
	// If Env is set, we do NOT want to overwrite it with Config.
	envSet := os.Getenv("MCPANY_MCP_LISTEN_ADDRESS") != ""

	if !cmd.Flags().Changed("mcp-listen-address") && !envSet && len(s.configPaths) > 0 {
		store := NewFileStore(fs, s.configPaths)
		store.SetSkipValidation(true)
		// We ignore errors here because we are only peeking for the listen address.
		// Real validation happens later in main.go or app.Run.
		// If we fail here, we prevent main.go from printing user-friendly errors for missing files.
		cfg, err := LoadResolvedConfig(context.Background(), store)
		if err == nil {
			if cfg.GetGlobalSettings().GetMcpListenAddress() != "" {
				mcpListenAddress = cfg.GetGlobalSettings().GetMcpListenAddress()
			}
			if len(cfg.GetGlobalSettings().GetMiddlewares()) > 0 {
				s.proto.SetMiddlewares(cfg.GetGlobalSettings().GetMiddlewares())
			}
		} else {
			// Log at debug level so we don't confuse the user if this was just a missing file that main.go will catch
			logging.GetLogger().Debug("Failed to peek config for listen address (this is expected if config is invalid or missing)", "error", err)
		}
	}
	s.proto.SetMcpListenAddress(mcpListenAddress)
	s.proto.SetLogLevel(s.LogLevel())
	s.proto.SetLogFormat(s.LogFormat())
// LogFormat returns the current log format as a protobuf enum.
//
// Summary: Retrieves the log format.
//
// Parameters:
//   - None.
//
// Returns:
// GRPCPort returns the gRPC port.
//
// Summary: Retrieves the gRPC port.
// MCPListenAddress returns the MCP listen address.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Retrieves the MCP listen address.
// MetricsListenAddress returns the metrics listen address.
//
// Summary: Retrieves the metrics listen address.
// Stdio returns whether stdio mode is enabled.
//
// Summary: Checks if stdio mode is enabled.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// ConfigPaths returns the paths to the configuration files.
//
// Summary: Retrieves configuration file paths.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// IsDebug returns whether debug mode is enabled.
//
// Summary: Checks if debug mode is enabled.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// LogFile returns the path to the log file.
//
// Summary: Retrieves the log file path.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// PersistentLog returns the path to the persistent log file used for hydration.
//
// Summary: Retrieves the persistent log file path.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// ShutdownTimeout returns the graceful shutdown timeout.
//
// Summary: Retrieves the shutdown timeout.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// APIKey returns the API key for the server.
//
// Summary: Retrieves the API key.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
// SetAPIKey sets the Global API key.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Sets the API key.
// SetMiddlewares sets the middlewares for the global settings.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Sets the middlewares.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Profiles returns the active profiles.
//
// Summary: Retrieves the active profiles.
//
// Parameters:
//   - None.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - []string: List of profile names.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// LogLevel returns the current log level as a protobuf enum.
//
// Summary: Retrieves the log level.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - configv1.GlobalSettings_LogLevel: The log level enum.
//
// Side Effects:
//   - Logs a warning if the log level is invalid.
// Errors:
//   - triggers relevant error states on failure.
func (s *Settings) LogLevel() configv1.GlobalSettings_LogLevel {
	if s.IsDebug() {
		return configv1.GlobalSettings_LOG_LEVEL_DEBUG
	}

	logLevel := strings.ToUpper(s.logLevel)
	// Handle "warning" as an alias for "WARN"
	if logLevel == "WARNING" {
		logLevel = "WARN"
	}

	key := "LOG_LEVEL_" + logLevel
	if val, ok := configv1.GlobalSettings_LogLevel_value[key]; ok {
		return configv1.GlobalSettings_LogLevel(val)
// DBPath returns the path to the SQLite database.
//
// Summary: Retrieves the database path.
// SetValues returns configuration values to override.
//
// Summary: Retrieves configuration override values.
// GetDbDsn returns the database DSN.
//
// Summary: Retrieves the database DSN.
// GetDbDriver returns the database driver.
//
// Summary: Retrieves the database driver.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Middlewares returns the configured middlewares.
//
// Summary: Retrieves the configured middlewares.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// GetDlp returns the DLP configuration.
//
// SetDlp sets the DLP configuration.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Sets the DLP configuration.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// GetOidc returns the OIDC configuration.
//
// Summary: Retrieves the OIDC configuration.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// GetProfileDefinitions returns the profile definitions.
//
// Summary: Retrieves the profile definitions.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// GithubAPIURL returns the GitHub API URL.
//
// Summary: Retrieves the GitHub API URL.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - None.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - string: The URL.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Side Effects:
//   - None.
// Errors:
//   - triggers relevant error states on failure.
func (s *Settings) GithubAPIURL() string {
	return s.proto.GetGithubApiUrl()
}

// getStringSlice is a helper function to get a string slice from viper.
// It handles the case where viper returns a slice with a single element
// containing comma-separated values (which happens with environment variables).
func getStringSlice(key string) []string {
	// Check the raw value to distinguish between a string (Env var) and a slice (YAML/JSON).
	raw := viper.Get(key)
	if val, ok := raw.(string); ok && val != "" {
		// It's a string, so it likely comes from an environment variable or flag.
		// We handle comma separation manually to avoid splitting by spaces within paths.
		if strings.Contains(val, ",") {
			parts := strings.Split(val, ",")
			var final []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					final = append(final, p)
				}
			}
			return final
		}
		return []string{strings.TrimSpace(val)}
	}

	// Fallback for slices (from config files) or empty values.
	res := viper.GetStringSlice(key)
	var final []string
	for _, item := range res {
		if strings.Contains(item, ",") {
			parts := strings.Split(item, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					final = append(final, p)
				}
			}
		} else {
			item = strings.TrimSpace(item)
			if item != "" {
				final = append(final, item)
			}
		}
	}
	return final
}
