// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package topology manages the network topology visualization and state.
package topology

import (
	"context"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	topologyv1 "github.com/mcpany/core/proto/topology/v1"
)

// Manager handles topology state tracking.
type Manager struct {
	mu              sync.RWMutex
	sessions        map[string]*SessionStats
	serviceRegistry serviceregistry.ServiceRegistryInterface
	toolManager     tool.ManagerInterface
}

// SessionStats contains statistics about a topology session.
type SessionStats struct {
	ID           string
	Metadata     map[string]string
	LastActive   time.Time
	RequestCount int64
	TotalLatency time.Duration
	ErrorCount   int64
}

// Stats aggregated metrics.
type Stats struct {
	TotalRequests int64
	AvgLatency    time.Duration
	ErrorRate     float64
}

// NewManager creates a new Topology Manager.
//
// registry is the registry.
// tm is the tm.
//
// Returns the result.
func NewManager(registry serviceregistry.ServiceRegistryInterface, tm tool.ManagerInterface) *Manager {
	return &Manager{
		sessions:        make(map[string]*SessionStats),
		serviceRegistry: registry,
		toolManager:     tm,
	}
}

// RecordActivity updates the session activity.
//
// sessionID is the sessionID.
// meta is the meta.
func (m *Manager) RecordActivity(sessionID string, meta map[string]interface{}, latency time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[sessionID]; !exists {
		// Convert generic map to string map for proto compatibility
		strMeta := make(map[string]string)
		for k, v := range meta {
			if s, ok := v.(string); ok {
				strMeta[k] = s
			}
		}

		m.sessions[sessionID] = &SessionStats{
			ID:       sessionID,
			Metadata: strMeta,
		}
	}
	m.sessions[sessionID].LastActive = time.Now()
	m.sessions[sessionID].RequestCount++
	m.sessions[sessionID].TotalLatency += latency
	if isError {
		m.sessions[sessionID].ErrorCount++
	}
}

// GetStats returns the aggregated stats.
func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalRequests int64
	var totalLatency time.Duration
	var totalErrors int64

	for _, session := range m.sessions {
		totalRequests += session.RequestCount
		totalLatency += session.TotalLatency
		totalErrors += session.ErrorCount
	}

	var avgLatency time.Duration
	var errorRate float64

	if totalRequests > 0 {
		avgLatency = time.Duration(int64(totalLatency) / totalRequests)
		errorRate = float64(totalErrors) / float64(totalRequests)
	}

	return Stats{
		TotalRequests: totalRequests,
		AvgLatency:    avgLatency,
		ErrorRate:     errorRate,
	}
}

// GetGraph generates the current topology graph.
//
// _ is an unused parameter.
//
// Returns the result.
func (m *Manager) GetGraph(_ context.Context) *topologyv1.Graph {
	m.mu.RLock()
	defer m.mu.RUnlock()

	coreNode := &topologyv1.Node{
		Id:       "mcp-core",
		Label:    "MCP Any",
		Type:     topologyv1.NodeType_NODE_TYPE_CORE,
		Status:   topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
		Children: []*topologyv1.Node{},
	}

	// Build Service -> Tool subtree
	services, err := m.serviceRegistry.GetAllServices()
	if err == nil {
		for _, svc := range services {
			svcNode := &topologyv1.Node{
				Id:     "svc-" + svc.GetName(),
				Label:  svc.GetName(),
				Type:   topologyv1.NodeType_NODE_TYPE_SERVICE,
				Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
			}
			if svc.GetDisable() {
				svcNode.Status = topologyv1.NodeStatus_NODE_STATUS_INACTIVE
			}

			// Add Tools
			tools := m.toolManager.ListTools()
			for _, t := range tools {
				if t.Tool().GetServiceId() == svc.GetName() {
					toolNode := &topologyv1.Node{
						Id:     "tool-" + t.Tool().GetName(),
						Label:  t.Tool().GetName(),
						Type:   topologyv1.NodeType_NODE_TYPE_TOOL,
						Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
					}

					// Mock API Call node
					apiNode := &topologyv1.Node{
						Id:     "api-" + t.Tool().GetName(),
						Label:  "POST /" + t.Tool().GetName(),
						Type:   topologyv1.NodeType_NODE_TYPE_API_CALL,
						Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
					}
					toolNode.Children = append(toolNode.Children, apiNode)

					svcNode.Children = append(svcNode.Children, toolNode)
				}
			}

			coreNode.Children = append(coreNode.Children, svcNode)
		}
	}

	// Add Middleware Nodes (Static or Dynamic)
	// For now, these are static infrastructure components in the pipeline
	middlewareNode := &topologyv1.Node{
		Id:     "middleware-pipeline",
		Label:  "Middleware Pipeline",
		Type:   topologyv1.NodeType_NODE_TYPE_MIDDLEWARE,
		Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
		Children: []*topologyv1.Node{
			{Id: "mw-auth", Label: "Authentication", Type: topologyv1.NodeType_NODE_TYPE_MIDDLEWARE, Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE},
			{Id: "mw-log", Label: "Logging", Type: topologyv1.NodeType_NODE_TYPE_MIDDLEWARE, Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE},
		},
	}
	coreNode.Children = append(coreNode.Children, middlewareNode)

	// Add Webhooks Node
	// This would ideally come from the WebhookManager
	webhookNode := &topologyv1.Node{
		Id:     "webhooks",
		Label:  "Webhooks",
		Type:   topologyv1.NodeType_NODE_TYPE_WEBHOOK,
		Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
		// Example configured webhook
		Children: []*topologyv1.Node{
			{Id: "wh-1", Label: "event-logger", Type: topologyv1.NodeType_NODE_TYPE_WEBHOOK, Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE},
		},
	}
	coreNode.Children = append(coreNode.Children, webhookNode)

	// Build Clients list from active sessions
	clients := make([]*topologyv1.Node, 0, len(m.sessions))
	for _, session := range m.sessions {
		// Filter out old sessions > 1 hour
		if time.Since(session.LastActive) > 1*time.Hour {
			continue
		}

		label := session.ID
		if name, ok := session.Metadata["userAgent"]; ok {
			label = name
		}

		clientNode := &topologyv1.Node{
			Id:     "client-" + session.ID,
			Label:  label,
			Type:   topologyv1.NodeType_NODE_TYPE_CLIENT,
			Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
			// Clients rely on UI to draw link to Core
		}
		clients = append(clients, clientNode)
	}

	return &topologyv1.Graph{
		Clients: clients,
		Core:    coreNode,
	}
}
