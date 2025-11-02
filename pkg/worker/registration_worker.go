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

package worker

import (
	"context"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/serviceregistry"
)

// ServiceRegistrationWorker is a background worker responsible for handling
// service registration requests. It listens for ServiceRegistrationRequest
// messages on the event bus, processes them using the service registry, and
// publishes the results as ServiceRegistrationResult messages.
type ServiceRegistrationWorker struct {
	bus             *bus.BusProvider
	serviceRegistry serviceregistry.ServiceRegistryInterface
}

// NewServiceRegistrationWorker creates a new ServiceRegistrationWorker.
//
// bus is the event bus used for receiving requests and publishing results.
// serviceRegistry is the registry that will handle the actual registration logic.
func NewServiceRegistrationWorker(bus *bus.BusProvider, serviceRegistry serviceregistry.ServiceRegistryInterface) *ServiceRegistrationWorker {
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
	log := logging.GetLogger().With("component", "ServiceRegistrationWorker")
	log.Info("Service registration worker started")

	requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](w.bus, bus.ServiceRegistrationRequestTopic)
	resultBus := bus.GetBus[*bus.ServiceRegistrationResult](w.bus, bus.ServiceRegistrationResultTopic)

	unsubscribe := requestBus.Subscribe(ctx, "request", func(req *bus.ServiceRegistrationRequest) {
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
			resultBus.Publish(ctx, req.CorrelationID(), res)
			return
		}

		serviceID, discoveredTools, discoveredResources, err := w.serviceRegistry.RegisterService(requestCtx, req.Config)

		res := &bus.ServiceRegistrationResult{
			ServiceKey:          serviceID,
			DiscoveredTools:     discoveredTools,
			DiscoveredResources: discoveredResources,
			Error:               err,
		}
		res.SetCorrelationID(req.CorrelationID())
		resultBus.Publish(ctx, req.CorrelationID(), res)
	})

	listRequestBus := bus.GetBus[*bus.ServiceListRequest](w.bus, bus.ServiceListRequestTopic)
	listResultBus := bus.GetBus[*bus.ServiceListResult](w.bus, bus.ServiceListResultTopic)

	listUnsubscribe := listRequestBus.Subscribe(ctx, "request", func(req *bus.ServiceListRequest) {
		log.Info("Received service list request", "correlationID", req.CorrelationID())
		services, err := w.serviceRegistry.GetAllServices()
		res := &bus.ServiceListResult{
			Services: services,
			Error:    err,
		}
		res.SetCorrelationID(req.CorrelationID())
		listResultBus.Publish(ctx, req.CorrelationID(), res)
	})

	go func() {
		<-ctx.Done()
		log.Info("Service registration worker stopping")
		unsubscribe()
		listUnsubscribe()
	}()
}
