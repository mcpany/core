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

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/health"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// UpstreamHealthCheck runs health checks for all configured upstream services.
func UpstreamHealthCheck(cmd *cobra.Command) error {
	fs := afero.NewOsFs()
	cfg := config.GlobalSettings()
	if err := cfg.Load(cmd, fs); err != nil {
		return err
	}

	store := config.NewFileStore(fs, cfg.ConfigPaths())
	serverConfig, err := config.LoadServices(store, "server")
	if err != nil {
		return fmt.Errorf("failed to load services from config: %w", err)
	}

	checker := health.NewChecker()

	for _, service := range serverConfig.GetUpstreamServices() {
		fmt.Fprintf(cmd.OutOrStdout(), "Checking service: %s\n", service.GetName())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := checker.Check(ctx, service); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  - Health check failed: %v\n", err)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  - Health check passed")
		}
	}

	return nil
}
