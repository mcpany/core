// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"sync"
	"time"

	config "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

// ServiceHealthProvider defines the interface for fetching service health.
type ServiceHealthProvider interface {
	GetAllServices() ([]*config.UpstreamServiceConfig, error)
}

// Point represents a snapshot of the system health at a point in time.
type Point struct {
	Timestamp       time.Time
	UptimePercentage float64
	HealthyServices  int32
	TotalServices    int32
}

// Recorder records historical health data.
type Recorder struct {
	mu       sync.RWMutex
	history  []Point
	maxPoints int
	provider ServiceHealthProvider
	interval time.Duration
}

// NewRecorder creates a new health recorder.
// It keeps history for 24 hours (assuming 1 minute interval = 1440 points).
func NewRecorder(provider ServiceHealthProvider) *Recorder {
	return &Recorder{
		history:   make([]Point, 0, 1440),
		maxPoints: 1440,
		provider:  provider,
		interval:  1 * time.Minute,
	}
}

// Start starts the background recording loop.
func (r *Recorder) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		// Initial record
		r.record(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.record(ctx)
			}
		}
	}()
}

func (r *Recorder) record(_ context.Context) {
	services, err := r.provider.GetAllServices()
	if err != nil {
		logging.GetLogger().Error("HealthRecorder: failed to get services", "error", err)
		return
	}

	//nolint:gosec // Safe conversion
	total := int32(len(services))
	healthy := int32(0)

	for _, svc := range services {
		// We consider a service healthy if it has no LastError
		if svc.GetLastError() == "" {
			healthy++
		}
	}

	var uptime float64
	if total > 0 {
		uptime = (float64(healthy) / float64(total)) * 100.0
	} else {
		// If no services, is it 100% uptime? Or 0?
		// System is running, so 100% "system" uptime, but 0 services.
		// Let's stick to 100% if no services are failing (because there are none).
		uptime = 100.0
	}

	point := Point{
		Timestamp:       time.Now().UTC(),
		UptimePercentage: uptime,
		HealthyServices:  healthy,
		TotalServices:    total,
	}

	r.mu.Lock()
	if len(r.history) >= r.maxPoints {
		// Remove oldest (shift left)
		// Efficient ring buffer would be better but slice copy is fine for 1440 items once a minute.
		r.history = r.history[1:]
	}
	r.history = append(r.history, point)
	r.mu.Unlock()
}

// GetHistory returns the recorded health history.
func (r *Recorder) GetHistory() []Point {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return copy
	result := make([]Point, len(r.history))
	copy(result, r.history)
	return result
}
