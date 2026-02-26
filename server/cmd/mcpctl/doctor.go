// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// DoctorRunner runs the doctor command checks.
//
// Summary: Runs the doctor command checks.
type DoctorRunner struct {
	Out        io.Writer
	Fs         afero.Fs
	HTTPClient *http.Client
}

// Run executes the doctor checks.
//
// Summary: Executes the doctor checks.
//
// Parameters:
//   - cmd (*cobra.Command): The cobra command to execute.
//   - _ ([]string): Ignored.
//
// Returns:
//   - error: An error if the operation fails.
//
// Side Effects:
//   - None.
func (r *DoctorRunner) Run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	const localhost = "localhost"
	_, _ = fmt.Fprintln(r.Out, "Running Doctor checks...")
	_, _ = fmt.Fprintln(r.Out, "========================")

	// 1. Check Configuration
	_, _ = fmt.Fprint(r.Out, "[ ] Checking Configuration... ")
	cfg := config.GlobalSettings()
	if err := cfg.Load(cmd, r.Fs); err != nil {
		_, _ = fmt.Fprintln(r.Out, "FAILED")
		return fmt.Errorf("configuration load failed: %w", err)
	}

	store := config.NewFileStore(r.Fs, cfg.ConfigPaths())
	configs, err := config.LoadServices(ctx, store, "server")
	if err != nil {
		// If loading fails, it might be due to no config files found, which might be okay or not.
		// But usually it returns empty list if just none found but path valid.
		_, _ = fmt.Fprintln(r.Out, "WARNING")
		_, _ = fmt.Fprintf(r.Out, "  -> Failed to load services: %v\n", err)
	} else {
		if validationErrors := config.Validate(ctx, configs, config.Server); len(validationErrors) > 0 {
			_, _ = fmt.Fprintln(r.Out, "FAILED")
			_, _ = fmt.Fprintln(r.Out, "  -> Validation errors:")
			for _, e := range validationErrors {
				_, _ = fmt.Fprintf(r.Out, "     - %s\n", e.Error())
			}
		} else {
			_, _ = fmt.Fprintln(r.Out, "OK")
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
	_, _ = fmt.Fprintf(r.Out, "[ ] Checking Server Connectivity (%s)... ", baseURL)

	// Check /health
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		_, _ = fmt.Fprintln(r.Out, "FAILED")
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		_, _ = fmt.Fprintln(r.Out, "FAILED")
		_, _ = fmt.Fprintf(r.Out, "  -> Could not connect to server: %v\n", err)
		_, _ = fmt.Fprintln(r.Out, "  -> Suggestion: Is the server running? Try 'mcpany run'")
		return nil // Don't error out, just report
	}
	// Close body immediately after check, or use a new variable.
	// We'll close it explicitly here since we don't need it after the check.
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = fmt.Fprintln(r.Out, "FAILED")
		_, _ = fmt.Fprintf(r.Out, "  -> Server returned status: %s\n", resp.Status)
		return nil
	}
	_, _ = fmt.Fprintln(r.Out, "OK")

	// 3. Check Deep Health (/doctor endpoint)
	_, _ = fmt.Fprint(r.Out, "[ ] Checking System Health... ")
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/doctor", nil)
	if err != nil {
		_, _ = fmt.Fprintln(r.Out, "WARNING")
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err = r.HTTPClient.Do(req)
	if err != nil {
		_, _ = fmt.Fprintln(r.Out, "WARNING")
		_, _ = fmt.Fprintf(r.Out, "  -> Could not query doctor endpoint: %v\n", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_, _ = fmt.Fprintln(r.Out, "WARNING")
		_, _ = fmt.Fprintf(r.Out, "  -> Doctor endpoint returned status: %s\n", resp.Status)
		return nil
	}

	var report health.DoctorReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		_, _ = fmt.Fprintln(r.Out, "WARNING")
		_, _ = fmt.Fprintf(r.Out, "  -> Failed to decode doctor report: %v\n", err)
		return nil
	}

	if report.Status == "healthy" || report.Status == "ok" {
		_, _ = fmt.Fprintln(r.Out, "OK")
	} else {
		_, _ = fmt.Fprintln(r.Out, "DEGRADED")
		_, _ = fmt.Fprintf(r.Out, "  -> Status: %s\n", report.Status)
	}

	// Print checks
	for name, result := range report.Checks {
		status := "OK"
		if result.Status != "ok" {
			status = "FAIL"
		}
		_, _ = fmt.Fprintf(r.Out, "    - %s: %s", name, status)
		if result.Latency != "" {
			_, _ = fmt.Fprintf(r.Out, " (%s)", result.Latency)
		}
		if result.Message != "" {
			_, _ = fmt.Fprintf(r.Out, " - %s", result.Message)
		}
		_, _ = fmt.Fprintln(r.Out)
	}

	return nil
}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &DoctorRunner{
				Out: cmd.OutOrStdout(),
				Fs:  afero.NewOsFs(),
				HTTPClient: &http.Client{
					Timeout: 2 * time.Second,
				},
			}
			return runner.Run(cmd, args)
		},
	}
}
