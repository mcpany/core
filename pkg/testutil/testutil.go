// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// WaitForServerReady polls the health check endpoint of the server until it gets a 200 OK response
// or the context is canceled.
func WaitForServerReady(ctx context.Context, addr string, timeout time.Duration) error {
	startTime := time.Now()
	healthCheckURL := fmt.Sprintf("http://%s/healthz", addr)

	for {
		if time.Since(startTime) > timeout {
			return fmt.Errorf("timed out waiting for server to be ready at %s", addr)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", healthCheckURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create health check request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil // Server is ready
		}
		if resp != nil {
			resp.Body.Close()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue polling
		}
	}
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}
