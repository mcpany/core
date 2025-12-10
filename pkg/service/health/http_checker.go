// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	config "github.com/mcpany/core/proto/config/v1"
)

// PerformHTTPCheck sends an HTTP request to the health check endpoint and
// verifies the response.
func PerformHTTPCheck(address string, hc *config.HttpHealthCheck) error {
	ctx, cancel := context.WithTimeout(context.Background(), hc.GetTimeout().AsDuration())
	defer cancel()

	url := hc.GetUrl()
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = fmt.Sprintf("%s%s", address, url)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if expectedCode := hc.GetExpectedCode(); expectedCode != 0 && resp.StatusCode != int(expectedCode) {
		return fmt.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, expectedCode)
	}

	if expectedBody := hc.GetExpectedResponseBodyContains(); expectedBody != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read health check response body: %w", err)
		}
		if !strings.Contains(string(body), expectedBody) {
			return fmt.Errorf("response body does not contain expected string %q", expectedBody)
		}
	}

	return nil
}
