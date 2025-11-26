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
	"time"

	"github.com/spf13/cobra"
)

// BindFlags binds the command line flags to viper.
func BindFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("mcp-listen-address", "50050", "MCP server's bind address.")
	cmd.PersistentFlags().StringSlice("config-path", []string{}, "Paths to configuration files or directories for pre-registering services.")
	cmd.Flags().String("grpc-port", "", "Port for the gRPC registration server.")
	cmd.Flags().Bool("stdio", false, "Enable stdio mode for JSON-RPC communication.")
	cmd.Flags().Bool("debug", false, "Enable debug logging.")
	cmd.Flags().String("log-level", "info", "Set the log level (debug, info, warn, error).")
	cmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Graceful shutdown timeout.")
	cmd.Flags().String("logfile", "", "Path to a file to write logs to.")
}
