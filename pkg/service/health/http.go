// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/proto/config/v1"
)

// HttpChecker implements the Checkable interface for HTTP services.
type HttpChecker struct {
	serviceID string
	config    *config.HttpHealthCheck
	client    *http.Client
}

// NewHttpChecker creates a new HttpChecker.
func NewHttpChecker(serviceID string, cfg *config.HttpHealthCheck, upstreamSvc upstream.Upstream) (*HttpChecker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("HTTP health check config is nil for service %s", serviceID)
	}
	if cfg.Url == "" {
		return nil, fmt.Errorf("HTTP health check URL is not specified for service %s", serviceID)
	}

	// Create a new HTTP client with a timeout matching the health check interval.
	// This ensures the client respects the overall timeout for the check.
	client := &http.Client{
		Timeout: upstream.DefaultTimeout,
	}
	if cfg.Timeout != nil {
		client.Timeout = cfg.Timeout.AsDuration()
	}

	return &HttpChecker{
		serviceID: serviceID,
		config:    cfg,
		client:    client,
	}, nil
}

// ID returns the unique identifier of the service.
func (c *HttpChecker) ID() string {
	return c.serviceID
}

// Interval returns the duration between health checks.
func (c *HttpChecker) Interval() time.Duration {
	if c.config.Interval != nil {
		return c.config.Interval.AsDuration()
	}
	// Return a default interval if not specified.
	return 15 * time.Second
}

// HealthCheck performs the health check for the HTTP service.
func (c *HttpChecker) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.Url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request for %s: %w", c.serviceID, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed for %s: %w", c.serviceID, err)
	}
	defer resp.Body.Close()

	expectedCode := int(c.config.ExpectedCode)
	if expectedCode == 0 {
		expectedCode = http.StatusOK
	}

	if resp.StatusCode != expectedCode {
		return fmt.Errorf("unexpected status code for %s: got %d, want %d", c.serviceID, resp.StatusCode, expectedCode)
	}

	if c.config.ExpectedResponseBodyContains != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body for %s: %w", c.serviceID, err)
		}
		if !strings.Contains(string(body), c.config.ExpectedResponseBodyContains) {
			return fmt.Errorf("response body for %s does not contain expected string", c.serviceID)
		}
	}

	return nil
}
