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

	"github.com/mcpxy/core/pkg/app"
	"github.com/mcpxy/core/pkg/appconsts"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appRun = app.Run

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   appconsts.Name,
		Short: "MCP-XY is a versatile proxy for backend services.",
		Run: func(cmd *cobra.Command, args []string) {
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

			if err := appRun(ctx, osFs, stdio, jsonrpcPort, registrationPort, configPaths); err != nil {
				log.Error("Application failed", "error", err)
			}
			log.Info("Shutdown complete.")
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpxy",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", appconsts.Name, appconsts.Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPXY")
	})

	rootCmd.Flags().String("jsonrpc-port", "50050", "Port for the JSON-RPC and HTTP registration server. Env: MCPXY_JSONRPC_PORT")
	rootCmd.Flags().String("grpc-port", "50051", "Port for the gRPC registration server. If not specified, gRPC registration is disabled. Env: MCPXY_GRPC_PORT")
	rootCmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication. Env: MCPXY_STDIO")
	rootCmd.Flags().StringSlice("config-paths", []string{}, "Paths to configuration files or directories for pre-registering services. Can be specified multiple times. Env: MCPXY_CONFIG_PATHS")
	rootCmd.Flags().Bool("debug", false, "Enable debug logging. Env: MCPXY_DEBUG")

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
		fmt.Println(err)
	}
}
