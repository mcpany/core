// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"strings"
	"time"

	v1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ServiceRegistryInterface defines the interface for the service registry.
type ServiceRegistryInterface interface {
	GetAllServices() ([]*configv1.UpstreamServiceConfig, error)
}

// Service implements the DiscoveryService gRPC interface.
type Service struct {
	v1.UnimplementedDiscoveryServiceServer
	registry ServiceRegistryInterface
}

// NewService creates a new DiscoveryService.
func NewService(registry ServiceRegistryInterface) *Service {
	return &Service{
		registry: registry,
	}
}

// GetIndexStatus returns the status of the tool index.
func (s *Service) GetIndexStatus(_ context.Context, _ *v1.GetIndexStatusRequest) (*v1.IndexStatus, error) {
	services, err := s.registry.GetAllServices()
	if err != nil {
		return nil, err
	}

	var totalTools int32
	for _, svc := range services {
		totalTools += int32(countTools(svc)) //nolint:gosec // Not expecting overflow here
	}

	return &v1.IndexStatus{
		TotalTools:   totalTools,
		IndexedTools: totalTools, // For now, we index everything in memory
		LastUpdated:  time.Now().Format(time.RFC3339),
	}, nil
}

// SearchTools searches for tools matching the query.
func (s *Service) SearchTools(_ context.Context, req *v1.SearchToolsRequest) (*v1.SearchToolsResponse, error) {
	services, err := s.registry.GetAllServices()
	if err != nil {
		return nil, err
	}

	var results []*v1.SearchResult
	query := strings.ToLower(req.Query)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	for _, svc := range services {
		tools := getTools(svc)
		for _, t := range tools {
			relevance := calculateRelevance(t, query)
			if relevance > 0 {
				results = append(results, &v1.SearchResult{
					Tool:        t,
					Relevance:   relevance,
					ServiceName: svc.GetName(),
				})
			}
		}
	}

	// Sort by relevance (descending)
	// Since we don't have a generic sort and want to keep it simple without extra deps,
	// let's do a simple bubble sort or just return as is if list is small.
	// For P0 MVP, basic filtering is enough, but sorting is better.
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Relevance < results[j].Relevance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return &v1.SearchToolsResponse{
		Results: results,
	}, nil
}

func countTools(svc *configv1.UpstreamServiceConfig) int {
	return len(getTools(svc))
}

func getTools(svc *configv1.UpstreamServiceConfig) []*configv1.ToolDefinition {
	if svc.GetCommandLineService() != nil {
		return svc.GetCommandLineService().GetTools()
	}
	if svc.GetHttpService() != nil {
		return svc.GetHttpService().GetTools()
	}
	if svc.GetGrpcService() != nil {
		return svc.GetGrpcService().GetTools()
	}
	if svc.GetOpenapiService() != nil {
		return svc.GetOpenapiService().GetTools()
	}
    // Add other service types as needed
	return nil
}

func calculateRelevance(tool *configv1.ToolDefinition, query string) float32 {
	name := strings.ToLower(tool.GetName())
	desc := strings.ToLower(tool.GetDescription())

	if query == "" {
		return 1.0
	}

	if name == query {
		return 1.0
	}
	if strings.Contains(name, query) {
		return 0.8
	}
	if strings.Contains(desc, query) {
		return 0.5
	}

	return 0.0
}
