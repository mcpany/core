/*
 * Copyright 2025 Author(s) of MCPX
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
	"github.com/mcpxy/mcpx/pkg/bus"
	"github.com/mcpxy/mcpx/pkg/logging"
	v1 "github.com/mcpxy/mcpx/proto/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegistrationServer implements the gRPC server for service registration.
// It handles gRPC requests for registering and managing upstream services by
// publishing messages to the event bus and waiting for the results from the
// corresponding workers.
type RegistrationServer struct {
	v1.UnimplementedRegistrationServiceServer
	bus *bus.BusProvider
}

// NewRegistrationServer creates a new RegistrationServer with the provided event
// bus.
//
// bus is the event bus used for communication with service registration workers.
func NewRegistrationServer(bus *bus.BusProvider) (*RegistrationServer, error) {
	if bus == nil {
		return nil, fmt.Errorf("bus is nil")
	}
	return &RegistrationServer{bus: bus}, nil
}

// RegisterService handles a gRPC request to register a new upstream service.
// It sends a registration request to the event bus and waits for a response from
// the registration worker. This process is asynchronous, allowing the server to
// remain responsive while the registration is in progress.
//
// ctx is the context for the gRPC call.
// req contains the configuration of the service to be registered.
func (s *RegistrationServer) RegisterService(ctx context.Context, req *v1.RegisterServiceRequest) (*v1.RegisterServiceResponse, error) {
	if req.GetConfig() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "config is required")
	}
	if req.GetConfig().GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "config.name is required")
	}

	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceRegistrationResult, 1)

	resultBus := bus.GetBus[*bus.ServiceRegistrationResult](s.bus, "service_registration_results")
	unsubscribe := resultBus.SubscribeOnce(correlationID, func(result *bus.ServiceRegistrationResult) {
		resultChan <- result
	})
	defer unsubscribe()

	requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](s.bus, "service_registration_requests")
	regReq := &bus.ServiceRegistrationRequest{
		Config: req.GetConfig(),
	}
	regReq.SetCorrelationID(correlationID)
	requestBus.Publish("request", regReq)

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
		)
		if len(result.DiscoveredTools) > 0 {
			for _, tool := range result.DiscoveredTools {
				log.InfoContext(ctx, "Discovered tool", "tool", tool)
			}
		}

		msg := fmt.Sprintf("service %s registered successfully with key %s", req.GetConfig().GetName(), result.ServiceKey)
		resp := &v1.RegisterServiceResponse{}
		resp.SetMessage(msg)
		resp.SetDiscoveredTools(result.DiscoveredTools)
		return resp, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded while waiting for service registration")
	case <-time.After(30 * time.Second): // Add a safety timeout
		return nil, status.Errorf(codes.DeadlineExceeded, "timed out waiting for service registration result")
	}
}

// UnregisterService is not yet implemented.
func (s *RegistrationServer) UnregisterService(ctx context.Context, req *v1.UnregisterServiceRequest) (*v1.UnregisterServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterService not implemented")
}

// InitiateOAuth2Flow is not yet implemented.
func (s *RegistrationServer) InitiateOAuth2Flow(ctx context.Context, req *v1.InitiateOAuth2FlowRequest) (*v1.InitiateOAuth2FlowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitiateOAuth2Flow not implemented")
}

// RegisterTools is not yet implemented.
func (s *RegistrationServer) RegisterTools(ctx context.Context, req *v1.RegisterToolsRequest) (*v1.RegisterToolsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterTools not implemented")
}

// GetServiceStatus is not yet implemented.
func (s *RegistrationServer) GetServiceStatus(ctx context.Context, req *v1.GetServiceStatusRequest) (*v1.GetServiceStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServiceStatus not implemented")
}
