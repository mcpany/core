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

package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Checker is responsible for running health checks for upstream services.
type Checker struct{}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{}
}

// Check runs a health check for the given service.
func (c *Checker) Check(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	healthCheck := service.GetHealthCheck()
	if healthCheck == nil {
		return nil
	}

	if hc := healthCheck.GetHttp(); hc != nil {
		return c.checkHTTP(ctx, hc)
	}
	if hc := healthCheck.GetGrpc(); hc != nil {
		return c.checkGRPC(ctx, hc)
	}
	if hc := healthCheck.GetCommandLine(); hc != nil {
		return c.checkCommandLine(ctx, hc)
	}
	return fmt.Errorf("unsupported health check type")
}

func (c *Checker) checkHTTP(ctx context.Context, hc *configv1.HttpHealthCheck) error {
	req, err := http.NewRequestWithContext(ctx, "GET", hc.GetUrl(), nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: time.Duration(hc.GetTimeout().GetSeconds()) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != int(hc.GetExpectedCode()) {
		return fmt.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, hc.GetExpectedCode())
	}

	return nil
}

func (c *Checker) checkGRPC(ctx context.Context, hc *configv1.GrpcHealthCheck) error {
	// TODO: Implement gRPC health check
	return nil
}

func (c *Checker) checkCommandLine(ctx context.Context, hc *configv1.CommandLineHealthCheck) error {
	// TODO: Implement command line health check
	return nil
}
