// Package topology manages the network topology visualization and state.

package topology

import (
	"context"
	"sync"
	"time"

	topologyv1 "github.com/mcpany/core/proto/topology/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
)

// activityEvent represents a single activity record event.
type activityEvent struct {
	SessionID string
	Meta      map[string]interface{}
	Latency   time.Duration
	IsError   bool
	ServiceID string
}

// Manager handles topology state tracking.
type Manager struct {
	mu              sync.RWMutex
	sessions        map[string]*SessionStats
	trafficHistory  map[int64]*MinuteStats // Unix timestamp (minute) -> stats
	serviceRegistry serviceregistry.ServiceRegistryInterface
	toolManager     tool.ManagerInterface

	activityCh chan activityEvent
	shutdownCh chan struct{}
}

// SessionStats contains statistics about a topology session.
type SessionStats struct {
	ID             string
	Metadata       map[string]string
	LastActive     time.Time
	RequestCount   int64
	TotalLatency   time.Duration
	ErrorCount     int64
	ServiceCounts  map[string]int64         // Per service request count
	ServiceErrors  map[string]int64         // Per service error count
	ServiceLatency map[string]time.Duration // Per service latency
}

// Stats aggregated metrics.
type Stats struct {
	TotalRequests int64
	AvgLatency    time.Duration
	ErrorRate     float64
}

// MinuteStats tracks stats for a single minute.
type MinuteStats struct {
	Requests     int64
	Errors       int64
	Latency      int64 // Total latency in ms
	ServiceStats map[string]*ServiceTrafficStats
}

// ServiceTrafficStats tracks stats for a single service in a minute.
type ServiceTrafficStats struct {
	Requests int64
	Errors   int64
	Latency  int64
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
	m := &Manager{
		sessions:        make(map[string]*SessionStats),
		trafficHistory:  make(map[int64]*MinuteStats),
		serviceRegistry: registry,
		toolManager:     tm,
		activityCh:      make(chan activityEvent, 1000), // Buffer of 1000 to handle bursts
		shutdownCh:      make(chan struct{}),
	}
	go m.processLoop()
	return m
}

// processLoop handles asynchronous activity recording to avoid locking the request path.
func (m *Manager) processLoop() {
	for {
		select {
		case <-m.shutdownCh:
			return
		case event := <-m.activityCh:
			m.handleActivity(event)
		}
	}
}

// handleActivity processes a single activity event.
func (m *Manager) handleActivity(event activityEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := event.SessionID
	meta := event.Meta
	latency := event.Latency
	isError := event.IsError
	serviceID := event.ServiceID

	if _, exists := m.sessions[sessionID]; !exists {
		// Convert generic map to string map for proto compatibility
		strMeta := make(map[string]string)
		for k, v := range meta {
			if s, ok := v.(string); ok {
				strMeta[k] = s
			}
		}

		m.sessions[sessionID] = &SessionStats{
			ID:             sessionID,
			Metadata:       strMeta,
			ServiceCounts:  make(map[string]int64),
			ServiceErrors:  make(map[string]int64),
			ServiceLatency: make(map[string]time.Duration),
		}
	}
	session := m.sessions[sessionID]
	session.LastActive = time.Now()
	session.RequestCount++
	session.TotalLatency += latency
	if isError {
		session.ErrorCount++
	}

	if serviceID != "" {
		if session.ServiceCounts == nil {
			session.ServiceCounts = make(map[string]int64)
			session.ServiceErrors = make(map[string]int64)
			session.ServiceLatency = make(map[string]time.Duration)
		}
		session.ServiceCounts[serviceID]++
		session.ServiceLatency[serviceID] += latency
		if isError {
			session.ServiceErrors[serviceID]++
		}
	}

	// Record traffic history
	now := time.Now().Truncate(time.Minute).Unix()
	if _, ok := m.trafficHistory[now]; !ok {
		m.trafficHistory[now] = &MinuteStats{
			ServiceStats: make(map[string]*ServiceTrafficStats),
		}
	}
	stats := m.trafficHistory[now]
	stats.Requests++
	stats.Latency += latency.Milliseconds()
	if isError {
		stats.Errors++
	}

	if serviceID != "" {
		if stats.ServiceStats == nil {
			stats.ServiceStats = make(map[string]*ServiceTrafficStats)
		}
		if _, ok := stats.ServiceStats[serviceID]; !ok {
			stats.ServiceStats[serviceID] = &ServiceTrafficStats{}
		}
		sStats := stats.ServiceStats[serviceID]
		sStats.Requests++
		sStats.Latency += latency.Milliseconds()
		if isError {
			sStats.Errors++
		}
	}

	// Cleanup old history (older than 24h) occasionally (every 100 requests roughly)
	if session.RequestCount%100 == 0 {
		cutoff := time.Now().Add(-24 * time.Hour).Unix()
		for t := range m.trafficHistory {
			if t < cutoff {
				delete(m.trafficHistory, t)
			}
		}
	}
}

// RecordActivity updates the session activity.
// ⚡ BOLT: Offloaded to asynchronous channel to prevent blocking the request path.
// Randomized Selection from Top 5 High-Impact Targets
//
// sessionID is the sessionID.
// meta is the meta.
// serviceID is the serviceID (optional).
func (m *Manager) RecordActivity(sessionID string, meta map[string]interface{}, latency time.Duration, isError bool, serviceID string) {
	// ⚡ BOLT: Shallow copy meta to prevent race conditions as map is passed by reference
	metaCopy := make(map[string]interface{}, len(meta))
	for k, v := range meta {
		metaCopy[k] = v
	}

	select {
	case m.activityCh <- activityEvent{
		SessionID: sessionID,
		Meta:      metaCopy,
		Latency:   latency,
		IsError:   isError,
		ServiceID: serviceID,
	}:
		// Successfully queued
	default:
		// Buffer full, drop event to prevent blocking
		logging.GetLogger().Warn("Topology manager activity buffer full, dropping event")
	}
}

// Close stops the background worker.
func (m *Manager) Close() {
	close(m.shutdownCh)
}

// GetStats returns the aggregated stats.
// serviceID is optional.
func (m *Manager) GetStats(serviceID string) Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalRequests int64
	var totalLatency time.Duration
	var totalErrors int64

	for _, session := range m.sessions {
		if serviceID != "" {
			if count, ok := session.ServiceCounts[serviceID]; ok {
				totalRequests += count
				totalLatency += session.ServiceLatency[serviceID]
				totalErrors += session.ServiceErrors[serviceID]
			}
		} else {
			totalRequests += session.RequestCount
			totalLatency += session.TotalLatency
			totalErrors += session.ErrorCount
		}
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
// serviceID is optional.
func (m *Manager) GetTrafficHistory(serviceID string) []TrafficPoint {
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
			if serviceID != "" && stats.ServiceStats != nil {
				if sStats, ok := stats.ServiceStats[serviceID]; ok {
					reqs = sStats.Requests
					errs = sStats.Errors
					lat = sStats.Latency
				}
			} else if serviceID == "" {
				reqs = stats.Requests
				errs = stats.Errors
				lat = stats.Latency
			}
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
	log := logging.GetLogger()
	log.Info("Seeding traffic history", "points", len(points))

	// Update a dummy session for GetStats to reflect seeded data
	m.sessions["seed-data"] = &SessionStats{
		ID:           "seed-data",
		LastActive:   now,
		RequestCount: 0,
		ErrorCount:   0,
		TotalLatency: 0,
	}

	for _, p := range points {
		// Parse time "HH:MM"
		t, err := time.Parse("15:04", p.Time)
		if err != nil {
			log.Error("Failed to parse seed time", "time", p.Time, "error", err)
			continue
		}
		// Adjust to today
		targetTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())

		// Handle day rollover: if the time is significantly in the future (e.g., > 1 hour), assume it was yesterday.
		if targetTime.After(now.Add(1 * time.Hour)) {
			targetTime = targetTime.Add(-24 * time.Hour)
		}

		// We assume seeded data is "Average Latency", so we multiply by requests to get total latency for storage
		m.trafficHistory[targetTime.Unix()] = &MinuteStats{
			Requests: p.Total,
			Errors:   p.Errors,
			Latency:  p.Latency * p.Total, // Reverse average
		}
		log.Info("Seeded point", "time", p.Time, "target_unix", targetTime.Unix(), "requests", p.Total)

		// Accumulate stats for the session
		m.sessions["seed-data"].RequestCount += p.Total
		m.sessions["seed-data"].ErrorCount += p.Errors
		m.sessions["seed-data"].TotalLatency += time.Duration(p.Latency*p.Total) * time.Millisecond
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

	coreNode := topologyv1.Node_builder{
		Id:     "mcp-core",
		Label:  "MCP Any",
		Type:   topologyv1.NodeType_NODE_TYPE_CORE,
		Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
	}.Build()

	// Build Service -> Tool subtree
	services, err := m.serviceRegistry.GetAllServices()
	if err == nil {
		// Get current traffic stats
		now := time.Now()
		minuteKey := now.Truncate(time.Minute).Unix()
		var currentStats *MinuteStats
		if stats, ok := m.trafficHistory[minuteKey]; ok {
			currentStats = stats
		}

		// Calculate elapsed seconds in current minute for QPS
		elapsedSeconds := float64(now.Second())
		if elapsedSeconds < 1 {
			elapsedSeconds = 1
		}

		// ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets.
		// Optimized nested loop to O(N+M) using a map.
		tools := m.toolManager.ListTools()
		toolsByService := make(map[string][]tool.Tool)
		for _, t := range tools {
			svcID := t.Tool().GetServiceId()
			toolsByService[svcID] = append(toolsByService[svcID], t)
		}

		for _, svc := range services {
			svcNode := topologyv1.Node_builder{
				Id:     "svc-" + svc.GetName(),
				Label:  svc.GetName(),
				Type:   topologyv1.NodeType_NODE_TYPE_SERVICE,
				Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
			}.Build()
			if svc.GetDisable() {
				svcNode.SetStatus(topologyv1.NodeStatus_NODE_STATUS_INACTIVE)
			}

			// Inject Metrics
			if currentStats != nil && currentStats.ServiceStats != nil {
				if sStats, ok := currentStats.ServiceStats[svc.GetName()]; ok {
					qps := float64(sStats.Requests) / elapsedSeconds
					errorRate := 0.0
					if sStats.Requests > 0 {
						errorRate = float64(sStats.Errors) / float64(sStats.Requests)
					}
					latency := 0.0
					if sStats.Requests > 0 {
						latency = float64(sStats.Latency) / float64(sStats.Requests)
					}

					svcNode.SetMetrics(topologyv1.NodeMetrics_builder{
						Qps:       qps,
						ErrorRate: errorRate,
						LatencyMs: latency,
					}.Build())

					// Dynamic Status
					if errorRate > 0.05 {
						svcNode.SetStatus(topologyv1.NodeStatus_NODE_STATUS_ERROR)
					}
				}
			}

			// Add Tools
			if svcTools, ok := toolsByService[svc.GetName()]; ok {
				for _, t := range svcTools {
					toolNode := topologyv1.Node_builder{
						Id:     "tool-" + t.Tool().GetName(),
						Label:  t.Tool().GetName(),
						Type:   topologyv1.NodeType_NODE_TYPE_TOOL,
						Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
					}.Build()

					// Mock API Call node
					apiNode := topologyv1.Node_builder{
						Id:     "api-" + t.Tool().GetName(),
						Label:  "POST /" + t.Tool().GetName(),
						Type:   topologyv1.NodeType_NODE_TYPE_API_CALL,
						Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
					}.Build()
					toolNode.SetChildren(append(toolNode.GetChildren(), apiNode))

					svcNode.SetChildren(append(svcNode.GetChildren(), toolNode))
				}
			}

			coreNode.SetChildren(append(coreNode.GetChildren(), svcNode))
		}
	}

	// Add Middleware Nodes (Static or Dynamic)
	// For now, these are static infrastructure components in the pipeline
	middlewareNode := topologyv1.Node_builder{
		Id:     "middleware-pipeline",
		Label:  "Middleware Pipeline",
		Type:   topologyv1.NodeType_NODE_TYPE_MIDDLEWARE,
		Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
		Children: []*topologyv1.Node{
			topologyv1.Node_builder{Id: "mw-auth", Label: "Authentication", Type: topologyv1.NodeType_NODE_TYPE_MIDDLEWARE, Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE}.Build(),
			topologyv1.Node_builder{Id: "mw-log", Label: "Logging", Type: topologyv1.NodeType_NODE_TYPE_MIDDLEWARE, Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE}.Build(),
		},
	}.Build()
	coreNode.SetChildren(append(coreNode.GetChildren(), middlewareNode))

	// Add Webhooks Node
	// This would ideally come from the WebhookManager
	webhookNode := topologyv1.Node_builder{
		Id:     "webhooks",
		Label:  "Webhooks",
		Type:   topologyv1.NodeType_NODE_TYPE_WEBHOOK,
		Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
		// Example configured webhook
		Children: []*topologyv1.Node{
			topologyv1.Node_builder{Id: "wh-1", Label: "event-logger", Type: topologyv1.NodeType_NODE_TYPE_WEBHOOK, Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE}.Build(),
		},
	}.Build()
	coreNode.SetChildren(append(coreNode.GetChildren(), webhookNode))

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

		clientNode := topologyv1.Node_builder{
			Id:     "client-" + session.ID,
			Label:  label,
			Type:   topologyv1.NodeType_NODE_TYPE_CLIENT,
			Status: topologyv1.NodeStatus_NODE_STATUS_ACTIVE,
			// Clients rely on UI to draw link to Core
		}.Build()
		clients = append(clients, clientNode)
	}

	return topologyv1.Graph_builder{
		Clients: clients,
		Core:    coreNode,
	}.Build()
}
