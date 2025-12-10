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
	"net/http"
	"time"

	configv1 "github.com.mcpany/core/proto/config/v1"
	"go.uber.org/zap"
)

func (c *checker) performHTTPCheck(ctx context.Context, serviceName, url string, check *configv1.HttpHealthCheck) {
	client := &http.Client{
		Timeout: check.GetTimeout().AsDuration(),
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("failed to create http request for health check", zap.String("service", serviceName), zap.Error(err))
		c.setStatus(serviceName, StatusUnhealthy)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Warn("health check failed", zap.String("service", serviceName), zap.Error(err))
		c.setStatus(serviceName, StatusUnhealthy)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != int(check.GetExpectedCode()) {
		c.logger.Warn("health check returned unexpected status code", zap.String("service", serviceName), zap.Int("status_code", resp.StatusCode))
		c.setStatus(serviceName, StatusUnhealthy)
		return
	}

	c.setStatus(serviceName, StatusHealthy)
}

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}
