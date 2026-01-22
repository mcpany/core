// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package topology manages the network topology visualization and state.
package topology

import (
	"context"
	"sync"
	"time"

	topologyv1 "github.com/mcpany/core/proto/topology/v1"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
)

// Manager handles topology state tracking.
type Manager struct {
	mu              sync.RWMutex
	sessions        map[string]*SessionStats
	trafficHistory  map[int64]*MinuteStats // Unix timestamp (minute) -> stats
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

// MinuteStats tracks stats for a single minute.
type MinuteStats struct {
	Requests int64
	Errors   int64
	Latency  int64 // Total latency in ms
}

// TrafficPoint represents a data point for the traffic chart.
type TrafficPoint struct {
	Time    string `json:"time"`
	Total   int64  `json:"requests"` // mapped to "requests" for UI
	Errors  int64  `json:"errors"`
	Latency int64  `json:"latency"`
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
		trafficHistory:  make(map[int64]*MinuteStats),
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

	// Record traffic history
	now := time.Now().Truncate(time.Minute).Unix()
	if _, ok := m.trafficHistory[now]; !ok {
		m.trafficHistory[now] = &MinuteStats{}
	}
	stats := m.trafficHistory[now]
	stats.Requests++
	stats.Latency += latency.Milliseconds()
	if isError {
		stats.Errors++
	}

	// Cleanup old history (older than 24h) occasionally (every 100 requests roughly)
	if m.sessions[sessionID].RequestCount%100 == 0 {
		cutoff := time.Now().Add(-24 * time.Hour).Unix()
		for t := range m.trafficHistory {
			if t < cutoff {
				delete(m.trafficHistory, t)
			}
		}
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

// GetTrafficHistory returns the traffic history for the last 24 hours.
// GetTrafficHistory returns the traffic history for the last 24 hours.
func (m *Manager) GetTrafficHistory() []TrafficPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	points := make([]TrafficPoint, 0, 60)
	now := time.Now()

	// Generate points for the last 60 minutes
	for i := 59; i >= 0; i-- {
		t := now.Add(time.Duration(-i) * time.Minute).Truncate(time.Minute)
		key := t.Unix()

		stats := m.trafficHistory[key]
		var reqs, errs, lat int64
		if stats != nil {
			reqs = stats.Requests
			errs = stats.Errors
			lat = stats.Latency
		}

		// Calculate avg latency for the point if needed, or just total.
		// UI `avgLatency` is calculated from total latency / total requests?
		// The mock data had `latency` as a value around 50-250.
		// If we return total latency, we should assume UI handles it?
		// UI code: `avgLatency = ... reduce(acc + cur.latency, 0) / length` -> This implies cur.latency is AVERAGE for that point.
		// So we should return Average Latency for that minute.

		avgLat := int64(0)
		if reqs > 0 {
			avgLat = lat / reqs
		}

		points = append(points, TrafficPoint{
			Time:    t.Format("15:04"),
			Total:   reqs,
			Errors:  errs,
			Latency: avgLat,
		})
	}
	return points
}

// SeedTrafficHistory allows seeding the traffic history with external data.
// This is primarily for testing and debugging purposes.
func (m *Manager) SeedTrafficHistory(points []TrafficPoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing history if needed, or just merge?
	// For seeding, usually we want to set state.
	// But since we store as map[int64]int64 (minute -> count), we need to parse the points.
	// The points are "HH:MM" -> total.
	// We should map them to today's timestamps.

	now := time.Now()
	// Clear current history
	m.trafficHistory = make(map[int64]*MinuteStats)

	for _, p := range points {
		// Parse time "HH:MM"
		t, err := time.Parse("15:04", p.Time)
		if err != nil {
			continue
		}
		// Adjust to today
		targetTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())

		// We assume seeded data is "Average Latency", so we multiply by requests to get total latency for storage
		m.trafficHistory[targetTime.Unix()] = &MinuteStats{
			Requests: p.Total,
			Errors:   p.Errors,
			Latency:  p.Latency * p.Total, // Reverse average
		}
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
