package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// newDoctorCmd creates the doctor command.
//
// This command checks the health of the MCP Any configuration and the running server.
// It performs configuration validation, server connectivity checks, and invokes the server's doctor endpoint.
//
// Returns:
//   - *cobra.Command: The configured doctor command.
func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the health of the MCP Any configuration and server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			const localhost = "localhost"
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Running Doctor checks...")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "========================")

			// 1. Check Configuration
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "[ ] Checking Configuration... ")
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "FAILED")
				return fmt.Errorf("configuration load failed: %w", err)
			}

			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			configs, err := config.LoadServices(ctx, store, "server")
			if err != nil {
				// If loading fails, it might be due to no config files found, which might be okay or not.
				// But usually it returns empty list if just none found but path valid.
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "WARNING")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Failed to load services: %v\n", err)
			} else {
				if validationErrors := config.Validate(ctx, configs, config.Server); len(validationErrors) > 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "FAILED")
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  -> Validation errors:")
					for _, e := range validationErrors {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "     - %s\n", e.Error())
					}
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "OK")
				}
			}

			// 2. Check Server Connectivity
			listenAddr := cfg.MCPListenAddress()
			if listenAddr == "" {
				listenAddr = "50050" // Default
			}

			// Normalize address for HTTP client
			host, port, err := net.SplitHostPort(listenAddr)
			if err != nil {
				// Assume it's just a port if split fails and it looks like a number
				if !strings.Contains(listenAddr, ":") {
					port = listenAddr
					host = localhost
				} else {
					// Handle :port case
					if strings.HasPrefix(listenAddr, ":") {
						port = strings.TrimPrefix(listenAddr, ":")
						host = localhost
					} else {
						// Fallback
						port = "50050"
						host = localhost
					}
				}
			}
			if host == "" || host == "0.0.0.0" {
				host = localhost
			}

			baseURL := fmt.Sprintf("http://%s:%s", host, port)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[ ] Checking Server Connectivity (%s)... ", baseURL)

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			// Check /health
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "FAILED")
				return fmt.Errorf("failed to create request: %w", err)
			}
			resp, err := client.Do(req)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "FAILED")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Could not connect to server: %v\n", err)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  -> Suggestion: Is the server running? Try 'mcpany run'")
				return nil // Don't error out, just report
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "FAILED")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Server returned status: %s\n", resp.Status)
				return nil
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "OK")

			// 3. Check Deep Health (/doctor endpoint)
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "[ ] Checking System Health... ")
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/doctor", nil)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "WARNING")
				return fmt.Errorf("failed to create request: %w", err)
			}
			resp, err = client.Do(req)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "WARNING")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Could not query doctor endpoint: %v\n", err)
				return nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "WARNING")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Doctor endpoint returned status: %s\n", resp.Status)
				return nil
			}

			var report health.DoctorReport
			if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "WARNING")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Failed to decode doctor report: %v\n", err)
				return nil
			}

			if report.Status == "healthy" || report.Status == "ok" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "OK")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "DEGRADED")
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  -> Status: %s\n", report.Status)
			}

			// Print checks
			for name, result := range report.Checks {
				status := "OK"
				if result.Status != "ok" {
					status = "FAIL"
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "    - %s: %s", name, status)
				if result.Latency != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), " (%s)", result.Latency)
				}
				if result.Message != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), " - %s", result.Message)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}
}
