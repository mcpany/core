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

package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
	})

	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server listen address. Env: MCPANY_MCP_LISTEN_ADDRESS")
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
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout. Env: MCPANY_SHUTDOWN_TIMEOUT")
	cmd.Flags().String("logfile", "", "Path to a file to write logs to. If not set, logs are written to stdout.")

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		fmt.Printf("Error binding command line flags: %v\n", err)
		os.Exit(1)
	}
}

// Config holds all the configuration for the application.
type Config struct {
	GRPC       *v1.GrpcUpstreamService
	MCP        *v1.McpUpstreamService
	Global     *v1.GlobalSettings
	shutdown   time.Duration
	configFile string
}

var (
	globalConfig *Config
	once         sync.Once
)

// GlobalConfig returns the global configuration instance.
func GlobalConfig() *Config {
	once.Do(func() {
		var err error
		globalConfig, err = LoadConfig()
		if err != nil {
			panic(err)
		}
	})
	return globalConfig
}

// GetBindAddress returns the bind address for the server.
// It prioritizes the config file over the command line flag if the flag is not explicitly set.
func (c *Config) GetBindAddress(cmd *cobra.Command, fs afero.Fs) (string, error) {
	bindAddress := viper.GetString("mcp-listen-address")
	configPaths := viper.GetStringSlice("config-path")

	if !cmd.Flags().Changed("mcp-listen-address") && len(configPaths) > 0 {
		store := NewFileStore(fs, configPaths)
		cfg, err := LoadServices(store, "server")
		if err != nil {
			return "", fmt.Errorf("failed to load services from config: %w", err)
		}
		if cfg.GetGlobalSettings().GetBindAddress() != "" {
			bindAddress = cfg.GetGlobalSettings().GetBindAddress()
		}
	}
	return bindAddress, nil
}

// GRPCPort returns the gRPC port.
func (c *Config) GRPCPort() string {
	if c.GRPC == nil {
		return ""
	}
	return c.GRPC.GetAddress()
}

// Stdio returns whether stdio mode is enabled.
func (c *Config) Stdio() bool {
	if c.MCP == nil {
		return false
	}
	return c.MCP.GetStdioConnection() != nil
}

// ConfigPaths returns the paths to the configuration files.
func (c *Config) ConfigPaths() []string {
	return viper.GetStringSlice("config-path")
}

// IsDebug returns whether debug mode is enabled.
func (c *Config) IsDebug() bool {
	return c.Global.GetLogLevel() == v1.GlobalSettings_DEBUG
}

// LogFile returns the path to the log file.
func (c *Config) LogFile() string {
	return c.Global.LogFile
}

// ShutdownTimeout returns the graceful shutdown timeout.
func (c *Config) ShutdownTimeout() time.Duration {
	return c.shutdown
}

// LogLevel returns the log level for the server.
func (c *Config) LogLevel() v1.GlobalSettings_LogLevel {
	return c.Global.GetLogLevel()
}
