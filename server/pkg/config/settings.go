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

// Settings defines the global configuration for the application.
//
// Summary. Represents a Settings.
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

// GlobalSettings provides globalsettings functionality.
//
// Summary: GlobalSettings.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func GlobalSettings() *Settings {
	once.Do(func() {
		globalSettings = &Settings{
			proto: configv1.GlobalSettings_builder{}.Build(),
		}
	})
	return globalSettings
}

// ToProto provides toproto functionality.
//
// Summary: ToProto.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) ToProto() *configv1.GlobalSettings {
	return s.proto
}

// Load provides load functionality.
//
// Summary: Load.
//
// Parameters.
//   - cmd: The parameter.
//   - fs: The parameter.
//
// Returns.
//   - result: The result.
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

// LogFormat provides logformat functionality.
//
// Summary: LogFormat.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) LogFormat() configv1.GlobalSettings_LogFormat {
	format := viper.GetString("log-format")
	key := "LOG_FORMAT_" + strings.ToUpper(format)
	if val, ok := configv1.GlobalSettings_LogFormat_value[key]; ok {
		return configv1.GlobalSettings_LogFormat(val)
	}
	return configv1.GlobalSettings_LOG_FORMAT_TEXT
}

// GRPCPort provides grpcport functionality.
//
// Summary: GRPCPort.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) GRPCPort() string {
	return s.grpcPort
}

// MCPListenAddress provides mcplistenaddress functionality.
//
// Summary: MCPListenAddress.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) MCPListenAddress() string {
	return s.proto.GetMcpListenAddress()
}

// MetricsListenAddress provides metricslistenaddress functionality.
//
// Summary: MetricsListenAddress.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) MetricsListenAddress() string {
	return viper.GetString("metrics-listen-address")
}

// Stdio provides stdio functionality.
//
// Summary: Stdio.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) Stdio() bool {
	return s.stdio
}

// ConfigPaths provides configpaths functionality.
//
// Summary: ConfigPaths.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) ConfigPaths() []string {
	return s.configPaths
}

// IsDebug provides isdebug functionality.
//
// Summary: IsDebug.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) IsDebug() bool {
	return s.debug
}

// LogFile provides logfile functionality.
//
// Summary: LogFile.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) LogFile() string {
	return s.logFile
}

// PersistentLog provides persistentlog functionality.
//
// Summary: PersistentLog.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) PersistentLog() string {
	return s.persistentLog
}

// ShutdownTimeout provides shutdowntimeout functionality.
//
// Summary: ShutdownTimeout.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) ShutdownTimeout() time.Duration {
	return s.shutdownTimeout
}

// APIKey provides apikey functionality.
//
// Summary: APIKey.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) APIKey() string {
	if s.proto.GetApiKey() != "" {
		return s.proto.GetApiKey()
	}
	return viper.GetString("api-key")
}

// SetAPIKey provides setapikey functionality.
//
// Summary: SetAPIKey.
//
// Parameters.
//   - key: The parameter.
//
// Returns.
//   - None.
func (s *Settings) SetAPIKey(key string) {
	s.proto.SetApiKey(key)
}

// SetMiddlewares provides setmiddlewares functionality.
//
// Summary: SetMiddlewares.
//
// Parameters.
//   - middlewares: The parameter.
//
// Returns.
//   - None.
func (s *Settings) SetMiddlewares(middlewares []*configv1.Middleware) {
	s.proto.SetMiddlewares(middlewares)
}

// Profiles provides profiles functionality.
//
// Summary: Profiles.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) Profiles() []string {
	if viper.IsSet("profiles") {
		return getStringSlice("profiles")
	}
	if len(s.profiles) == 0 {
		return []string{"default"}
	}
	return s.profiles
}

// LogLevel provides loglevel functionality.
//
// Summary: LogLevel.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// DBPath provides dbpath functionality.
//
// Summary: DBPath.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) DBPath() string {
	return s.dbPath
}

// SetValues provides setvalues functionality.
//
// Summary: SetValues.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) SetValues() []string {
	return s.setValues
}

// GetDbDsn provides getdbdsn functionality.
//
// Summary: GetDbDsn.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) GetDbDsn() string {
	return s.proto.GetDbDsn()
}

// GetDbDriver provides getdbdriver functionality.
//
// Summary: GetDbDriver.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) GetDbDriver() string {
	return s.proto.GetDbDriver()
}

// Middlewares provides middlewares functionality.
//
// Summary: Middlewares.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) Middlewares() []*configv1.Middleware {
	return s.proto.GetMiddlewares()
}

// GetDlp provides getdlp functionality.
//
// Summary: GetDlp.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) GetDlp() *configv1.DLPConfig {
	return s.proto.GetDlp()
}

// SetDlp provides setdlp functionality.
//
// Summary: SetDlp.
//
// Parameters.
//   - dlp: The parameter.
//
// Returns.
//   - None.
func (s *Settings) SetDlp(dlp *configv1.DLPConfig) {
	s.proto.SetDlp(dlp)
}

// GetOidc provides getoidc functionality.
//
// Summary: GetOidc.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) GetOidc() *configv1.OIDCConfig {
	return s.proto.GetOidc()
}

// GetProfileDefinitions provides getprofiledefinitions functionality.
//
// Summary: GetProfileDefinitions.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Settings) GetProfileDefinitions() []*configv1.ProfileDefinition {
	return s.proto.GetProfileDefinitions()
}

// GithubAPIURL provides githubapiurl functionality.
//
// Summary: GithubAPIURL.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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
