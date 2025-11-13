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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitViper(rootCmd *cobra.Command) {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MCPANY")
	})

	if err := viper.BindPFlag("jsonrpc-port", rootCmd.PersistentFlags().Lookup("jsonrpc-port")); err != nil {
		fmt.Printf("Error binding jsonrpc-port flag: %v\n", err)
	}
	if err := viper.BindPFlag("config-path", rootCmd.PersistentFlags().Lookup("config-path")); err != nil {
		fmt.Printf("Error binding config-path flag: %v\n", err)
	}

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		fmt.Printf("Error binding command line flags: %v\n", err)
	}
}

func Load() *Config {
	return &Config{
		JSONRPCPort:      viper.GetString("jsonrpc-port"),
		Stdio:            viper.GetBool("stdio"),
		ConfigPaths:      viper.GetStringSlice("config-path"),
		Debug:            viper.GetBool("debug"),
		LogFile:          viper.GetString("logfile"),
		ShutdownTimeout:  viper.GetDuration("shutdown-timeout"),
		RegistrationPort: viper.GetString("grpc-port"),
	}
}
