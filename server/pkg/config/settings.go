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

// Settings represents the public Settings entity.
//
// Summary: Defines the structured data model representing a .
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

// GlobalSettings serves as a public interface for interacting with GlobalSettings.
//
// Summary: Global the settings appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func GlobalSettings() *Settings {
	once.Do(func() {
		globalSettings = &Settings{
			proto: configv1.GlobalSettings_builder{}.Build(),
		}
	})
	return globalSettings
}

// ToProto serves as a public interface for interacting with ToProto.
//
// Summary: To the proto appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) ToProto() *configv1.GlobalSettings {
	return s.proto
}

// Load serves as a public interface for interacting with Load.
//
// Summary: Load the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) Load(cmd *cobra.Command, fs afero.Fs) error {
	s.cmd = cmd
	s.fs = fs

	s.grpcPort = viper.GetString("grpc-port")
	s.stdio = viper.GetBool("stdio") // Corrected from "std"
	// Bind config paths
	s.configPaths = getStringSlice("config-path")
	if len(s.configPaths) == 0 && viper.IsSet("config") {
		s.configPaths = getStringSlice("config")
	}
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

// LogFormat serves as a public interface for interacting with LogFormat.
//
// Summary: Log the format appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) LogFormat() configv1.GlobalSettings_LogFormat {
	format := viper.GetString("log-format")
	key := "LOG_FORMAT_" + strings.ToUpper(format)
	if val, ok := configv1.GlobalSettings_LogFormat_value[key]; ok {
		return configv1.GlobalSettings_LogFormat(val)
	}
	return configv1.GlobalSettings_LOG_FORMAT_TEXT
}

// GRPCPort serves as a public interface for interacting with GRPCPort.
//
// Summary: Grpc the port appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GRPCPort() string {
	return s.grpcPort
}

// MCPListenAddress serves as a public interface for interacting with MCPListenAddress.
//
// Summary: Mcp the listen address appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) MCPListenAddress() string {
	return s.proto.GetMcpListenAddress()
}

// MetricsListenAddress serves as a public interface for interacting with MetricsListenAddress.
//
// Summary: Metrics the listen address appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) MetricsListenAddress() string {
	return viper.GetString("metrics-listen-address")
}

// Stdio serves as a public interface for interacting with Stdio.
//
// Summary: Stdio the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) Stdio() bool {
	return s.stdio
}

// ConfigPaths serves as a public interface for interacting with ConfigPaths.
//
// Summary: Config the paths appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) ConfigPaths() []string {
	return s.configPaths
}

// IsDebug serves as a public interface for interacting with IsDebug.
//
// Summary: Checks condition indicating whether the target is debug.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) IsDebug() bool {
	return s.debug
}

// LogFile serves as a public interface for interacting with LogFile.
//
// Summary: Log the file appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) LogFile() string {
	return s.logFile
}

// PersistentLog serves as a public interface for interacting with PersistentLog.
//
// Summary: Persistent the log appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) PersistentLog() string {
	return s.persistentLog
}

// ShutdownTimeout serves as a public interface for interacting with ShutdownTimeout.
//
// Summary: Shutdown the timeout appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) ShutdownTimeout() time.Duration {
	return s.shutdownTimeout
}

// APIKey serves as a public interface for interacting with APIKey.
//
// Summary: Api the key appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) APIKey() string {
	if s.proto.GetApiKey() != "" {
		return s.proto.GetApiKey()
	}
	return viper.GetString("api-key")
}

// SetAPIKey serves as a public interface for interacting with SetAPIKey.
//
// Summary: Set the api key appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) SetAPIKey(key string) {
	s.proto.SetApiKey(key)
}

// SetMiddlewares serves as a public interface for interacting with SetMiddlewares.
//
// Summary: Set the middlewares appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) SetMiddlewares(middlewares []*configv1.Middleware) {
	s.proto.SetMiddlewares(middlewares)
}

// Profiles serves as a public interface for interacting with Profiles.
//
// Summary: Profiles the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) Profiles() []string {
	if viper.IsSet("profiles") {
		return getStringSlice("profiles")
	}
	if len(s.profiles) == 0 {
		return []string{"default"}
	}
	return s.profiles
}

// LogLevel serves as a public interface for interacting with LogLevel.
//
// Summary: Log the level appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DBPath serves as a public interface for interacting with DBPath.
//
// Summary: Db the path appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) DBPath() string {
	return s.dbPath
}

// SetValues serves as a public interface for interacting with SetValues.
//
// Summary: Set the values appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) SetValues() []string {
	return s.setValues
}

// GetDbDsn serves as a public interface for interacting with GetDbDsn.
//
// Summary: Fetches and returns the underlying db dsn from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GetDbDsn() string {
	return s.proto.GetDbDsn()
}

// GetDbDriver serves as a public interface for interacting with GetDbDriver.
//
// Summary: Fetches and returns the underlying db driver from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GetDbDriver() string {
	return s.proto.GetDbDriver()
}

// Middlewares serves as a public interface for interacting with Middlewares.
//
// Summary: Middlewares the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) Middlewares() []*configv1.Middleware {
	return s.proto.GetMiddlewares()
}

// GetDlp serves as a public interface for interacting with GetDlp.
//
// Summary: Fetches and returns the underlying dlp from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GetDlp() *configv1.DLPConfig {
	return s.proto.GetDlp()
}

// SetDlp serves as a public interface for interacting with SetDlp.
//
// Summary: Set the dlp appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) SetDlp(dlp *configv1.DLPConfig) {
	s.proto.SetDlp(dlp)
}

// GetSso serves as a public interface for interacting with GetSso.
//
// Summary: Fetches and returns the underlying sso from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GetSso() *configv1.SSOConfig {
	return s.proto.GetSso()
}

// SetSso serves as a public interface for interacting with SetSso.
//
// Summary: Set the sso appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) SetSso(sso *configv1.SSOConfig) {
	s.proto.SetSso(sso)
}

// GetOidc serves as a public interface for interacting with GetOidc.
//
// Summary: Fetches and returns the underlying oidc from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GetOidc() *configv1.OIDCConfig {
	return s.proto.GetOidc()
}

// GetProfileDefinitions serves as a public interface for interacting with GetProfileDefinitions.
//
// Summary: Fetches and returns the underlying profile definitions from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Settings) GetProfileDefinitions() []*configv1.ProfileDefinition {
	return s.proto.GetProfileDefinitions()
}

// GithubAPIURL serves as a public interface for interacting with GithubAPIURL.
//
// Summary: Github the apiurl appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
