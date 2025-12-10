// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"log"
	"time"

	"github.com/mcpany/core/pkg/healthstatus"
	"github.com/mcpany/core/pkg/serviceregistry"
	config "github.com/mcpany/core/proto/config/v1"
)

// Executor manages the lifecycle of health checks for all services.
type Executor struct {
	registry serviceregistry.ServiceRegistryInterface
	ctx      context.Context
	cancel   context.CancelFunc
	tickers  map[string]*time.Ticker
}

// NewExecutor creates a new HealthCheckExecutor.
func NewExecutor(ctx context.Context, registry serviceregistry.ServiceRegistryInterface) *Executor {
	ctx, cancel := context.WithCancel(ctx)
	return &Executor{
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,
		tickers:  make(map[string]*time.Ticker),
	}
}

// Start initiates the health checking process for all registered services.
func (e *Executor) Start() {
	services, err := e.registry.GetAllServices()
	if err != nil {
		log.Printf("ERROR: could not retrieve services for health checks: %v", err)
		return
	}

	for _, serviceConfig := range services {
		e.RegisterHealthCheck(serviceConfig)
	}
}

// RegisterHealthCheck starts a health check for a single service if configured.
func (e *Executor) RegisterHealthCheck(serviceConfig *config.UpstreamServiceConfig) {
	serviceID := serviceConfig.GetId()
	var interval time.Duration
	var checker func() error

	switch sc := serviceConfig.ServiceConfig.(type) {
	case *config.UpstreamServiceConfig_HttpService:
		if hc := sc.HttpService.GetHealthCheck(); hc != nil {
			interval = hc.GetInterval().AsDuration()
			checker = func() error {
				return PerformHTTPCheck(sc.HttpService.GetAddress(), hc)
			}
		}
	// case *config.UpstreamServiceConfig_GrpcService:
	// 	if hc := sc.GrpcService.GetHealthCheck(); hc != nil {
	// 		// To be implemented
	// 	}
	default:
		// This service type does not have a health check defined.
		return
	}

	if checker == nil || interval <= 0 {
		// No valid health check config, so we assume it's healthy.
		e.registry.SetHealthStatus(serviceID, healthstatus.HEALTHY)
		return
	}

	ticker := time.NewTicker(interval)
	e.tickers[serviceID] = ticker

	go func() {
		defer ticker.Stop()
		// Perform an initial check immediately.
		e.performCheck(serviceID, checker)

		for {
			select {
			case <-ticker.C:
				e.performCheck(serviceID, checker)
			case <-e.ctx.Done():
				return
			}
		}
	}()
}

func (e *Executor) performCheck(serviceID string, checker func() error) {
	if err := checker(); err != nil {
		log.Printf("Health check failed for service %s: %v", serviceID, err)
		e.registry.SetHealthStatus(serviceID, healthstatus.UNHEALTHY)
	} else {
		log.Printf("Health check passed for service %s", serviceID)
		e.registry.SetHealthStatus(serviceID, healthstatus.HEALTHY)
	}
}

// Stop terminates all running health checks.
func (e *Executor) Stop() {
	e.cancel()
}
