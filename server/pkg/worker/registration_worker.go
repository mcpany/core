// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package worker provides background worker functionality.
package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/util"
)

// ServiceRegistrationWorker is a background worker responsible for handling
// service registration requests. It listens for ServiceRegistrationRequest
// messages on the event bus, processes them using the service registry, and
// publishes the results as ServiceRegistrationResult messages.
type ServiceRegistrationWorker struct {
	bus             *bus.Provider
	serviceRegistry serviceregistry.ServiceRegistryInterface
	wg              sync.WaitGroup
	retryDelay      time.Duration
}

// NewServiceRegistrationWorker creates a new ServiceRegistrationWorker.
//
// Parameters:
//   - bus: The event bus used for receiving requests and publishing results.
//   - serviceRegistry: The registry that will handle the actual registration logic.
//
// Returns:
//   - *ServiceRegistrationWorker: A new service registration worker.
func NewServiceRegistrationWorker(bus *bus.Provider, serviceRegistry serviceregistry.ServiceRegistryInterface) *ServiceRegistrationWorker {
	return &ServiceRegistrationWorker{
		bus:             bus,
		serviceRegistry: serviceRegistry,
		retryDelay:      5 * time.Second,
	}
}

// SetRetryDelay sets the retry delay for failed registrations.
//
// Parameters:
//   - d: The duration to wait before retrying.
func (w *ServiceRegistrationWorker) SetRetryDelay(d time.Duration) {
	w.retryDelay = d
}

// Start launches the worker in a new goroutine. It subscribes to service
// registration requests on the event bus and will continue to process them
// until the provided context is canceled.
//
// Parameters:
//   - ctx: The context that controls the lifecycle of the worker.
func (w *ServiceRegistrationWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	log := logging.GetLogger().With("component", "ServiceRegistrationWorker")
	log.Info("Service registration worker started")

	requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](w.bus, bus.ServiceRegistrationRequestTopic)
	if err != nil {
		log.Error("Failed to get service registration request bus", "error", err)
		return
	}
	resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](w.bus, bus.ServiceRegistrationResultTopic)
	if err != nil {
		log.Error("Failed to get service registration result bus", "error", err)
		return
	}

	unsubscribe := requestBus.Subscribe(ctx, "request", func(req *bus.ServiceRegistrationRequest) {
		// Process registration in a separate goroutine to prevent blocking other registrations
		go func() {
			start := time.Now()
			metrics.IncrCounter([]string{"worker", "registration", "request", "total"}, 1)
			defer metrics.MeasureSince([]string{"worker", "registration", "request", "latency"}, start)

			// Panic recovery for registration request processing
			defer func() {
				if r := recover(); r != nil {
					metrics.IncrCounter([]string{"worker", "registration", "request", "panics"}, 1)
					log.Error("Panic during service registration", "panic", r, "stack", string(debug.Stack()))
					// Publish failure result so caller is not hanging if waiting (though this is fire-and-forget mostly)
					res := &bus.ServiceRegistrationResult{
						Error: fmt.Errorf("panic during registration: %v", r),
					}
					res.SetCorrelationID(req.CorrelationID())
					if err := resultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
						log.Error("Failed to publish panic result", "error", err)
					}
				}
			}()

			log.Info("Received service registration request", "correlationID", req.CorrelationID())

			requestCtx := req.Context
			if requestCtx == nil {
				requestCtx = context.Background()
			}

			if req.Config.GetDisable() {
				log.Info("Unregistering disabled service", "service", req.Config.GetName())
				err := w.serviceRegistry.UnregisterService(requestCtx, req.Config.GetName())
				res := &bus.ServiceRegistrationResult{
					Error: err,
				}
				res.SetCorrelationID(req.CorrelationID())
				if err := resultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
					log.Error("Failed to publish unregister result", "error", err)
				}
				return
			}

			// Apply timeout from resilience config if present
			if resilience := req.Config.GetResilience(); resilience != nil && resilience.GetTimeout() != nil {
				var cancel context.CancelFunc
				timeoutDuration := resilience.GetTimeout().AsDuration()
				// If timeout is very small, we might want to enforce a minimum?
				// For now, trust the config.
				if timeoutDuration > 0 {
					requestCtx, cancel = context.WithTimeout(requestCtx, timeoutDuration)
					defer cancel()
				}
			}

			serviceID, discoveredTools, discoveredResources, err := w.serviceRegistry.RegisterService(requestCtx, req.Config)

			res := &bus.ServiceRegistrationResult{
				ServiceKey:          serviceID,
				DiscoveredTools:     discoveredTools,
				DiscoveredResources: discoveredResources,
				Error:               err,
			}
			if err != nil {
				log.Error("Failed to register service", "service", req.Config.GetName(), "error", err)
				metrics.IncrCounter([]string{"worker", "registration", "request", "error"}, 1)

				// Schedule a retry
				// Simple fixed delay for now. In a robust system, we would track retry counts and apply backoff.
				// Since we don't have a place to store retry count in the request without modifying proto,
				// we just retry indefinitely every X seconds (configured via retryDelay) until success or cancellation.
				retryDelay := w.retryDelay
				log.Info("Scheduling retry for service registration", "service", req.Config.GetName(), "delay", retryDelay)

				go func() {
					select {
					case <-ctx.Done():
						return
					case <-time.After(retryDelay):
						if err := requestBus.Publish(ctx, "request", req); err != nil {
							log.Error("Failed to publish retry request", "service", req.Config.GetName(), "error", err)
						}
					}
				}()
			} else {
				log.Info("Successfully registered service", "service", req.Config.GetName(), "tools_count", len(discoveredTools), "resources_count", len(discoveredResources))
				metrics.IncrCounter([]string{"worker", "registration", "request", "success"}, 1)
			}
			res.SetCorrelationID(req.CorrelationID())
			if err := resultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
				log.Error("Failed to publish registration result", "error", err)
			}
		}()
	})

	listRequestBus, err := bus.GetBus[*bus.ServiceListRequest](w.bus, bus.ServiceListRequestTopic)
	if err != nil {
		log.Error("Failed to get service list request bus", "error", err)
		return
	}
	listResultBus, err := bus.GetBus[*bus.ServiceListResult](w.bus, bus.ServiceListResultTopic)
	if err != nil {
		log.Error("Failed to get service list result bus", "error", err)
		return
	}

	listUnsubscribe := listRequestBus.Subscribe(ctx, "request", func(req *bus.ServiceListRequest) {
		// Panic recovery for list request processing
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic during service list", "panic", r, "stack", string(debug.Stack()))
				res := &bus.ServiceListResult{
					Error: fmt.Errorf("panic during service list: %v", r),
				}
				res.SetCorrelationID(req.CorrelationID())
				if err := listResultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
					log.Error("Failed to publish list panic result", "error", err)
				}
			}
		}()

		log.Info("Received service list request", "correlationID", req.CorrelationID())
		services, err := w.serviceRegistry.GetAllServices()
		res := &bus.ServiceListResult{
			Services: services,
			Error:    err,
		}
		res.SetCorrelationID(req.CorrelationID())
		if err := listResultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
			log.Error("Failed to publish list result", "error", err)
		}
	})

	getRequestBus, err := bus.GetBus[*bus.ServiceGetRequest](w.bus, bus.ServiceGetRequestTopic)
	if err != nil {
		log.Error("Failed to get service get request bus", "error", err)
		return
	}
	getResultBus, err := bus.GetBus[*bus.ServiceGetResult](w.bus, bus.ServiceGetResultTopic)
	if err != nil {
		log.Error("Failed to get service get result bus", "error", err)
		return
	}

	getUnsubscribe := getRequestBus.Subscribe(ctx, "request", func(req *bus.ServiceGetRequest) {
		// Panic recovery for get request processing
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic during service get", "panic", r, "stack", string(debug.Stack()))
				res := &bus.ServiceGetResult{
					Error: fmt.Errorf("panic during service get: %v", r),
				}
				res.SetCorrelationID(req.CorrelationID())
				if err := getResultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
					log.Error("Failed to publish get panic result", "error", err)
				}
			}
		}()

		log.Info("Received service get request", "correlationID", req.CorrelationID(), "serviceName", req.ServiceName)
		// We use GetServiceConfig because GetServiceInfo returns Tool/Prompt info, but we want the config.
		// However, GetServiceConfig expects the serviceID (sanitized name).
		// The request might contain the display name or the ID.
		// We try as provided first.
		service, ok := w.serviceRegistry.GetServiceConfig(req.ServiceName)
		if !ok {
			// Try sanitizing the name
			sanitized, err := util.SanitizeServiceName(req.ServiceName)
			if err == nil {
				service, ok = w.serviceRegistry.GetServiceConfig(sanitized)
			}
		}

		var err error
		if !ok {
			err = fmt.Errorf("service %q not found", req.ServiceName)
		}

		res := &bus.ServiceGetResult{
			Service: service,
			Error:   err,
		}

		res.SetCorrelationID(req.CorrelationID())
		if err := getResultBus.Publish(ctx, req.CorrelationID(), res); err != nil {
			log.Error("Failed to publish get result", "error", err)
		}
	})

	go func() {
		defer w.wg.Done()
		<-ctx.Done()
		log.Info("Service registration worker stopping")
		unsubscribe()
		listUnsubscribe()
		getUnsubscribe()
	}()
}

// Stop waits for the worker to stop.
func (w *ServiceRegistrationWorker) Stop() {
	w.wg.Wait()
}
