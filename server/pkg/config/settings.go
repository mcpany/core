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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Summary: Settings defines the global configuration for the application. Represents a Settings.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
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
	shutdownTimeout time.Duration
	profiles        []string
	dbPath          string
	setValues       []string
	fs              afero.Fs
	cmd             *cobra.Command
}

var (
	globalSettings *Settings
	once           sync.Once
)

// Summary: GlobalSettings returns the singleton instance of the global settings. Retrieves the global settings singleton.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Settings: The resulting *Settings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GlobalSettings() *Settings {
	once.Do(func() {
		globalSettings = &Settings{
			proto: configv1.GlobalSettings_builder{}.Build(),
		}
	})
	return globalSettings
}

// Summary: ToProto returns the underlying GlobalSettings protobuf message. Converts the settings to a protobuf message.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.GlobalSettings: The resulting *configv1.GlobalSettings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) ToProto() *configv1.GlobalSettings {
	return s.proto
}

// Summary: Load initializes the global settings from the command line and config files. Loads configuration from flags and files.
//
// Parameters:
//   - cmd (*cobra.Command): The cmd parameter.
//   - fs (afero.Fs): The fs parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
	s.proto.SetApiKey(s.APIKey())

	// Set DB settings from config file if available, otherwise viper defaults (flags/env)
	if s.proto.GetDbDsn() == "" {
		s.proto.SetDbDsn(viper.GetString("db-dsn"))
	}
	if s.proto.GetDbDriver() == "" {
		s.proto.SetDbDriver(viper.GetString("db-driver"))
	}

	return nil
}

// Summary: LogFormat returns the current log format as a protobuf enum. Retrieves the log format.
//
// Parameters:
//   - None.
//
// Returns:
//   - configv1.GlobalSettings_LogFormat: The resulting configv1.GlobalSettings_LogFormat.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) LogFormat() configv1.GlobalSettings_LogFormat {
	format := viper.GetString("log-format")
	key := "LOG_FORMAT_" + strings.ToUpper(format)
	if val, ok := configv1.GlobalSettings_LogFormat_value[key]; ok {
		return configv1.GlobalSettings_LogFormat(val)
	}
	return configv1.GlobalSettings_LOG_FORMAT_TEXT
}

// Summary: GRPCPort returns the gRPC port. Retrieves the gRPC port.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) GRPCPort() string {
	return s.grpcPort
}

// Summary: MCPListenAddress returns the MCP listen address. Retrieves the MCP listen address.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) MCPListenAddress() string {
	return s.proto.GetMcpListenAddress()
}

// Summary: MetricsListenAddress returns the metrics listen address. Retrieves the metrics listen address.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) MetricsListenAddress() string {
	return viper.GetString("metrics-listen-address")
}

// Summary: Stdio returns whether stdio mode is enabled. Checks if stdio mode is enabled.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) Stdio() bool {
	return s.stdio
}

// Summary: ConfigPaths returns the paths to the configuration files. Retrieves configuration file paths.
//
// Parameters:
//   - None.
//
// Returns:
//   - []string: The resulting []string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) ConfigPaths() []string {
	return s.configPaths
}

// Summary: IsDebug returns whether debug mode is enabled. Checks if debug mode is enabled.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) IsDebug() bool {
	return s.debug
}

// Summary: LogFile returns the path to the log file. Retrieves the log file path.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) LogFile() string {
	return s.logFile
}

// Summary: PersistentLog returns the path to the persistent log file used for hydration. Retrieves the persistent log file path.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) PersistentLog() string {
	return s.persistentLog
}

// Summary: ShutdownTimeout returns the graceful shutdown timeout. Retrieves the shutdown timeout.
//
// Parameters:
//   - None.
//
// Returns:
//   - time.Duration: The resulting time.Duration.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) ShutdownTimeout() time.Duration {
	return s.shutdownTimeout
}

// Summary: APIKey returns the API key for the server. Retrieves the API key.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) APIKey() string {
	if s.proto.GetApiKey() != "" {
		return s.proto.GetApiKey()
	}
	return viper.GetString("api-key")
}

// Summary: SetAPIKey sets the Global API key. Sets the API key.
//
// Parameters:
//   - key (string): The key parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) SetAPIKey(key string) {
	s.proto.SetApiKey(key)
}

// Summary: SetMiddlewares sets the middlewares for the global settings. Sets the middlewares.
//
// Parameters:
//   - middlewares ([]*configv1.Middleware): The middlewares parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) SetMiddlewares(middlewares []*configv1.Middleware) {
	s.proto.SetMiddlewares(middlewares)
}

// Summary: Profiles returns the active profiles. Retrieves the active profiles.
//
// Parameters:
//   - None.
//
// Returns:
//   - []string: The resulting []string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) Profiles() []string {
	if viper.IsSet("profiles") {
		return getStringSlice("profiles")
	}
	if len(s.profiles) == 0 {
		return []string{"default"}
	}
	return s.profiles
}

// Summary: LogLevel returns the current log level as a protobuf enum. Retrieves the log level.
//
// Parameters:
//   - None.
//
// Returns:
//   - configv1.GlobalSettings_LogLevel: The resulting configv1.GlobalSettings_LogLevel.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
	}

	if s.logLevel != "" {
		logging.GetLogger().Warn(
			fmt.Sprintf(
				"Invalid log level specified: '%s'. Defaulting to INFO.",
				s.logLevel,
			),
		)
	}
	return configv1.GlobalSettings_LOG_LEVEL_INFO
}

// Summary: DBPath returns the path to the SQLite database. Retrieves the database path.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) DBPath() string {
	return s.dbPath
}

// Summary: SetValues returns configuration values to override. Retrieves configuration override values.
//
// Parameters:
//   - None.
//
// Returns:
//   - []string: The resulting []string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) SetValues() []string {
	return s.setValues
}

// Summary: GetDbDsn returns the database DSN. Retrieves the database DSN.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) GetDbDsn() string {
	return s.proto.GetDbDsn()
}

// Summary: GetDbDriver returns the database driver. Retrieves the database driver.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) GetDbDriver() string {
	return s.proto.GetDbDriver()
}

// Summary: Middlewares returns the configured middlewares. Retrieves the configured middlewares.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*configv1.Middleware: The resulting []*configv1.Middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) Middlewares() []*configv1.Middleware {
	return s.proto.GetMiddlewares()
}

// Summary: GetDlp returns the DLP configuration. Retrieves the DLP configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.DLPConfig: The resulting *configv1.DLPConfig.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) GetDlp() *configv1.DLPConfig {
	return s.proto.GetDlp()
}

// Summary: SetDlp sets the DLP configuration. Sets the DLP configuration.
//
// Parameters:
//   - dlp (*configv1.DLPConfig): The dlp parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) SetDlp(dlp *configv1.DLPConfig) {
	s.proto.SetDlp(dlp)
}

// Summary: GetOidc returns the OIDC configuration. Retrieves the OIDC configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.OIDCConfig: The resulting *configv1.OIDCConfig.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) GetOidc() *configv1.OIDCConfig {
	return s.proto.GetOidc()
}

// Summary: GetProfileDefinitions returns the profile definitions. Retrieves the profile definitions.
//
// Parameters:
//   - None.
//
// Returns:
//   - []*configv1.ProfileDefinition: The resulting []*configv1.ProfileDefinition.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Settings) GetProfileDefinitions() []*configv1.ProfileDefinition {
	return s.proto.GetProfileDefinitions()
}

// Summary: GithubAPIURL returns the GitHub API URL. Retrieves the GitHub API URL.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
