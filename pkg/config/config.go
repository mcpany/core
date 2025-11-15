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

// Config holds all configuration for the application.
type Config struct {
	BindAddress     string
	GRPCPort        string
	Stdio           bool
	ConfigPaths     []string
	Debug           bool
	LogFile         string
	ShutdownTimeout time.Duration
	fileConfig      *v1.McpAnyServerConfig
}

// GlobalSettings is a singleton instance of Config.
var (
	GlobalSettings *Config
	once           sync.Once
)

// Options holds the options for loading the configuration.
type Options struct {
	Cmd *cobra.Command
	Fs  afero.Fs
}

// Load initializes the global configuration.
func Load(opts Options) error {
	var err error
	once.Do(func() {
		bindAddress := viper.GetString("mcp-listen-address")
		configPaths := viper.GetStringSlice("config-path")

		var fileConfig *v1.McpAnyServerConfig
		if !opts.Cmd.Flags().Changed("mcp-listen-address") && len(configPaths) > 0 {
			store := NewFileStore(opts.Fs, configPaths)
			fileConfig, err = LoadServices(store, "server")
			if err != nil {
				err = fmt.Errorf("failed to load services from config: %w", err)
				return
			}
			if fileConfig.GetGlobalSettings().GetBindAddress() != "" {
				bindAddress = fileConfig.GetGlobalSettings().GetBindAddress()
			}
		}

		GlobalSettings = &Config{
			BindAddress:     bindAddress,
			GRPCPort:        viper.GetString("grpc-port"),
			Stdio:           viper.GetBool("stdio"),
			ConfigPaths:     configPaths,
			Debug:           viper.GetBool("debug"),
			LogFile:         viper.GetString("logfile"),
			ShutdownTimeout: viper.GetDuration("shutdown-timeout"),
			fileConfig:      fileConfig,
		}
	})
	return err
}

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
	})

	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server's bind address. Env: MCPANY_MCP_LISTEN_ADDRESS")
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

// GetLogLevel returns the log level for the server.
func (c *Config) GetLogLevel() v1.GlobalSettings_LogLevel {
	if c.Debug {
		return v1.GlobalSettings_DEBUG
	}
	return v1.GlobalSettings_INFO
}

// GetFileConfig returns the configuration loaded from files.
func (c *Config) GetFileConfig() *v1.McpAnyServerConfig {
	return c.fileConfig
}

// ResetForTesting resets the singleton for testing purposes.
func ResetForTesting() {
	once = sync.Once{}
	GlobalSettings = nil
}
