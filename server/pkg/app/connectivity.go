// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// performConnectivityCheck runs a dynamic connectivity check for the given service.
// It returns (valid, details, latency, error).
func performConnectivityCheck(ctx context.Context, svc *configv1.UpstreamServiceConfig) (bool, string, time.Duration, error) {
	var checkErr error
	var checkDetails string
	start := time.Now()

	// HTTP & GraphQL
	if httpSvc := svc.GetHttpService(); httpSvc != nil {
		checkErr = checkURLReachability(ctx, httpSvc.GetAddress())
		checkDetails = "HTTP reachability check failed"
	} else if gqlSvc := svc.GetGraphqlService(); gqlSvc != nil {
		checkErr = checkURLReachability(ctx, gqlSvc.GetAddress())
		checkDetails = "GraphQL reachability check failed"
	} else if fsSvc := svc.GetFilesystemService(); fsSvc != nil {
		// Filesystem check
		for _, path := range fsSvc.GetRootPaths() {
			if err := checkFilesystemAccess(path); err != nil {
				checkErr = err
				checkDetails = fmt.Sprintf("Filesystem path check failed for %s", path)
				break
			}
		}
	} else if cmdSvc := svc.GetCommandLineService(); cmdSvc != nil {
		// Command check
		checkErr = checkCommandAvailability(cmdSvc.GetCommand(), cmdSvc.GetWorkingDirectory())
		checkDetails = "Command availability check failed"
	} else if mcpSvc := svc.GetMcpService(); mcpSvc != nil {
		// MCP Remote check (if stdio, check command; if http, check url)
		switch mcpSvc.WhichConnectionType() {
		case configv1.McpUpstreamService_StdioConnection_case:
			stdio := mcpSvc.GetStdioConnection()
			if stdio != nil {
				checkErr = checkCommandAvailability(stdio.GetCommand(), stdio.GetWorkingDirectory())
				checkDetails = "MCP Stdio command check failed"
			}
		case configv1.McpUpstreamService_HttpConnection_case:
			httpConn := mcpSvc.GetHttpConnection()
			if httpConn != nil {
				checkErr = checkURLReachability(ctx, httpConn.GetHttpAddress())
				checkDetails = "MCP HTTP reachability check failed"
			}
		}
	}

	if checkErr != nil {
		return false, checkDetails, time.Since(start), checkErr
	}

	return true, "Connectivity check passed", time.Since(start), nil
}

func checkURLReachability(ctx context.Context, urlStr string) error {
	client := util.NewSafeHTTPClient()
	client.Timeout = 5 * time.Second

	// Try HEAD first
	req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to GET if HEAD is not supported (Method Not Allowed) or fails
		req, err = http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to reach %s: %w", urlStr, err)
		}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusUnauthorized {
		// We treat 401/403 as "reachable but requires auth", which is fine for basic connectivity check (auth check is deeper).
		// But 404 or 500 might indicate issues.
		// Actually, for validation, maybe we should be strict?
		// Let's just warn if it's 5xx. 404 might be valid if it's a base URL.
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server returned error status: %s", resp.Status)
		}
	}
	return nil
}

func checkFilesystemAccess(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}
	// We allow both files and directories, so existence is sufficient validation for now.
	return nil
}

func checkCommandAvailability(command string, workDir string) error {
	if command == "" {
		return fmt.Errorf("command is empty")
	}

	// If absolute path, check existence
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err != nil {
			return fmt.Errorf("executable not found at %s", command)
		}
	} else {
		// Look in PATH
		if _, err := exec.LookPath(command); err != nil {
			return fmt.Errorf("command %s not found in PATH", command)
		}
	}

	// Check working directory if provided
	if workDir != "" {
		info, err := os.Stat(workDir)
		if err != nil {
			return fmt.Errorf("working directory not found: %s", workDir)
		}
		if !info.IsDir() {
			return fmt.Errorf("working directory path is not a directory: %s", workDir)
		}
	}

	return nil
}
