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

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/util"
)

// ServiceRegistrationWorker is a background worker responsible for handling
// service registration requests. It listens for ServiceRegistrationRequest
// messages on the event bus, processes them using the service registry, and
// publishes the results as ServiceRegistrationResult messages.
type ServiceRegistrationWorker struct {
	bus             *bus.Provider
	serviceRegistry serviceregistry.ServiceRegistryInterface
	wg              sync.WaitGroup
}

// NewServiceRegistrationWorker creates a new ServiceRegistrationWorker.
//
// bus is the event bus used for receiving requests and publishing results.
// serviceRegistry is the registry that will handle the actual registration logic.
func NewServiceRegistrationWorker(bus *bus.Provider, serviceRegistry serviceregistry.ServiceRegistryInterface) *ServiceRegistrationWorker {
	return &ServiceRegistrationWorker{
		bus:             bus,
		serviceRegistry: serviceRegistry,
	}
}

// Start launches the worker in a new goroutine. It subscribes to service
// registration requests on the event bus and will continue to process them
// until the provided context is canceled.
//
// ctx is the context that controls the lifecycle of the worker.
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
					_ = resultBus.Publish(ctx, req.CorrelationID(), res)
				}
			}()

			log.Info("Received service registration request", "correlationID", req.CorrelationID(), "service", req.Config.GetName(), "retry", req.RetryCount)

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
				_ = resultBus.Publish(ctx, req.CorrelationID(), res)
				return
			}

			serviceID, discoveredTools, discoveredResources, err := w.serviceRegistry.RegisterService(requestCtx, req.Config)

			if err != nil {
				metrics.IncrCounter([]string{"worker", "registration", "request", "error"}, 1)
				log.Error("Failed to register service", "service", req.Config.GetName(), "error", err)

				// Retry logic
				// We retry up to 10 times, with exponential backoff capped at 30 seconds.
				const MaxRetries = 10
				if req.RetryCount < MaxRetries {
					req.RetryCount++
					delay := time.Duration(1<<uint(req.RetryCount)) * time.Second
					if delay > 30*time.Second {
						delay = 30 * time.Second
					}
					log.Info("Retrying service registration", "service", req.Config.GetName(), "retry", req.RetryCount, "delay", delay)

					time.Sleep(delay)
					if err := requestBus.Publish(ctx, "request", req); err != nil {
						log.Error("Failed to republish registration request", "error", err)
					}
					return // Return early, don't publish failure result yet
				}
			} else {
				metrics.IncrCounter([]string{"worker", "registration", "request", "success"}, 1)
			}

			res := &bus.ServiceRegistrationResult{
				ServiceKey:          serviceID,
				DiscoveredTools:     discoveredTools,
				DiscoveredResources: discoveredResources,
				Error:               err,
			}
			res.SetCorrelationID(req.CorrelationID())
			_ = resultBus.Publish(ctx, req.CorrelationID(), res)
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
				_ = listResultBus.Publish(ctx, req.CorrelationID(), res)
			}
		}()

		log.Info("Received service list request", "correlationID", req.CorrelationID())
		services, err := w.serviceRegistry.GetAllServices()
		res := &bus.ServiceListResult{
			Services: services,
			Error:    err,
		}
		res.SetCorrelationID(req.CorrelationID())
		_ = listResultBus.Publish(ctx, req.CorrelationID(), res)
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
				_ = getResultBus.Publish(ctx, req.CorrelationID(), res)
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
		_ = getResultBus.Publish(ctx, req.CorrelationID(), res)
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
