// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/mcpany/core/server/pkg/health"
)

// servicesHealthCheck checks the health of all registered upstream services.
func (a *Application) servicesHealthCheck(ctx context.Context) health.CheckResult {
	start := time.Now()

	if a.ServiceRegistry == nil {
		return health.CheckResult{
			Status:  "ok",
			Message: "Service registry not initialized",
			Latency: time.Since(start).String(),
		}
	}

	services, err := a.ServiceRegistry.GetAllServices()
	if err != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("Failed to list services: %v", err),
			Latency: time.Since(start).String(),
		}
	}

	if len(services) == 0 {
		return health.CheckResult{
			Status:  "ok",
			Message: "No services registered",
			Latency: time.Since(start).String(),
		}
	}

	var wg sync.WaitGroup
	results := make([]doctor.CheckResult, len(services))

	for i := range services {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Use a derived context with timeout
			checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			res := doctor.CheckService(checkCtx, services[index])
			res.ServiceName = services[index].GetName()
			results[index] = res
		}(i)
	}

	wg.Wait()

	// Aggregate results
	var failed []string
	status := "ok"

	for _, res := range results {
		if res.Status != doctor.StatusOk && res.Status != doctor.StatusSkipped {
			status = "degraded"
			msg := res.Message
			if msg == "" && res.Error != nil {
				msg = res.Error.Error()
			}
			// Use service name if set, otherwise try to find it
			name := res.ServiceName
			if name == "" {
				name = "unknown"
			}
			failed = append(failed, fmt.Sprintf("%s: %s", name, msg))
		}
	}

	message := fmt.Sprintf("%d services healthy", len(services))
	if len(failed) > 0 {
		message = strings.Join(failed, "; ")
	}

	return health.CheckResult{
		Status:  status,
		Message: message,
		Latency: time.Since(start).String(),
	}
}
