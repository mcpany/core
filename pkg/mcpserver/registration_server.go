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

package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	v1 "github.com/mcpany/core/proto/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegistrationServer implements the gRPC server for service registration. It
// handles gRPC requests for registering and managing upstream services by
// publishing messages to the event bus and waiting for the results from the
// corresponding workers. This decouples the gRPC server from the core service
// registration logic, allowing for a more modular and scalable architecture.
type RegistrationServer struct {
	v1.UnimplementedRegistrationServiceServer
	bus *bus.BusProvider
}

// NewRegistrationServerHook is a test hook for overriding the creation of a
// RegistrationServer.
var NewRegistrationServerHook func(bus interface{}) (*RegistrationServer, error)

// NewRegistrationServer creates a new RegistrationServer initialized with the
// event bus.
//
// The bus is used for communicating with the service registration workers,
// allowing for an asynchronous, decoupled registration process.
//
// Parameters:
//   - bus: The event bus used for communication.
//
// Returns a new instance of the RegistrationServer or an error if the bus is
// nil.
func NewRegistrationServer(bus *bus.BusProvider) (*RegistrationServer, error) {
	if NewRegistrationServerHook != nil {
		// The type assertion is safe because this is a test-only hook.
		return NewRegistrationServerHook(bus)
	}
	if bus == nil {
		return nil, fmt.Errorf("bus is nil")
	}
	return &RegistrationServer{bus: bus}, nil
}

// RegisterService handles a gRPC request to register a new upstream service.
// It sends a registration request to the event bus and waits for a response
// from the registration worker. This process is asynchronous, allowing the
// server to remain responsive while the registration is in progress.
//
// A correlation ID is used to match the request with the corresponding result
// from the worker. The method waits for the result, with a timeout, and returns
// the registration details, including any discovered tools.
//
// Parameters:
//   - ctx: The context for the gRPC call.
//   - req: The request containing the configuration of the service to be
//     registered.
//
// Returns a response with the registration status and discovered tools, or an
// error if the registration fails or times out.
func (s *RegistrationServer) RegisterService(ctx context.Context, req *v1.RegisterServiceRequest) (*v1.RegisterServiceResponse, error) {
	if req.GetConfig() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "config is required")
	}
	if req.GetConfig().GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "config.name is required")
	}

	if err := config.ValidateOrError(req.GetConfig()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
	}

	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceRegistrationResult, 1)

	resultBus := bus.GetBus[*bus.ServiceRegistrationResult](s.bus, "service_registration_results")
	unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ServiceRegistrationResult) {
		resultChan <- result
	})
	defer unsubscribe()

	requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](s.bus, "service_registration_requests")
	regReq := &bus.ServiceRegistrationRequest{
		Config: req.GetConfig(),
	}
	regReq.SetCorrelationID(correlationID)
	_ = requestBus.Publish(ctx, "request", regReq)

	// Wait for the result, respecting the context's deadline
	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "failed to register service: %v", result.Error)
		}

		log := logging.GetLogger()
		log.InfoContext(ctx, "Service registered via bus",
			"service_name", req.GetConfig().GetName(),
			"service_key", result.ServiceKey,
			"discovered_tools_count", len(result.DiscoveredTools),
			"discovered_resources_count", len(result.DiscoveredResources),
		)
		if len(result.DiscoveredTools) > 0 {
			for _, tool := range result.DiscoveredTools {
				log.InfoContext(ctx, "Discovered tool", "tool", tool)
			}
		}
		if len(result.DiscoveredResources) > 0 {
			for _, resource := range result.DiscoveredResources {
				log.InfoContext(ctx, "Discovered resource", "resource", resource)
			}
		}

		msg := fmt.Sprintf("service %s registered successfully with key %s", req.GetConfig().GetName(), result.ServiceKey)
		resp := &v1.RegisterServiceResponse{}
		resp.SetMessage(msg)
		resp.SetDiscoveredTools(result.DiscoveredTools)
		resp.SetServiceKey(result.ServiceKey)
		resp.SetDiscoveredResources(result.DiscoveredResources)
		return resp, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded while waiting for service registration")
	case <-time.After(30 * time.Second): // Add a safety timeout
		return nil, status.Errorf(codes.DeadlineExceeded, "timed out waiting for service registration result")
	}
}

// UnregisterService is not yet implemented. It is intended to handle the
// unregistration of a service.
func (s *RegistrationServer) UnregisterService(ctx context.Context, req *v1.UnregisterServiceRequest) (*v1.UnregisterServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterService not implemented")
}

// InitiateOAuth2Flow is not yet implemented. It is intended to handle the
// initiation of an OAuth2 flow for a service.
func (s *RegistrationServer) InitiateOAuth2Flow(ctx context.Context, req *v1.InitiateOAuth2FlowRequest) (*v1.InitiateOAuth2FlowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitiateOAuth2Flow not implemented")
}

// RegisterTools is not yet implemented. It is intended to handle the
// registration of tools for a service.
func (s *RegistrationServer) RegisterTools(ctx context.Context, req *v1.RegisterToolsRequest) (*v1.RegisterToolsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterTools not implemented")
}

// GetServiceStatus is not yet implemented. It is intended to handle requests
// for the status of a service.
func (s *RegistrationServer) GetServiceStatus(ctx context.Context, req *v1.GetServiceStatusRequest) (*v1.GetServiceStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServiceStatus not implemented")
}

func (s *RegistrationServer) mustEmbedUnimplementedRegistrationServiceServer() {}

func (s *RegistrationServer) ListServices(ctx context.Context, req *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceListResult, 1)

	resultBus := bus.GetBus[*bus.ServiceListResult](s.bus, "service_list_results")
	unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ServiceListResult) {
		resultChan <- result
	})
	defer unsubscribe()

	requestBus := bus.GetBus[*bus.ServiceListRequest](s.bus, "service_list_requests")
	listReq := &bus.ServiceListRequest{}
	listReq.SetCorrelationID(correlationID)
	_ = requestBus.Publish(ctx, "request", listReq)

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "failed to list services: %v", result.Error)
		}
		resp := &v1.ListServicesResponse{}
		resp.SetServices(result.Services)
		return resp, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded while waiting for service list")
	case <-time.After(30 * time.Second):
		return nil, status.Errorf(codes.DeadlineExceeded, "timed out waiting for service list result")
	}
}
