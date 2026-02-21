// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/topology"
)

// handleDebugSeed provides an endpoint to seed various data for testing/demo purposes.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type seedRequest struct {
			Type    string `json:"type"`
			Count   int    `json:"count"`
			Service string `json:"service"`
			Clear   bool   `json:"clear"`
		}

		var req seedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		switch req.Type {
		case "traffic":
			if a.TopologyManager == nil {
				http.Error(w, "topology manager not enabled", http.StatusBadRequest)
				return
			}
			seedTraffic(a.TopologyManager, req.Service, req.Count)

		case "logs":
			seedLogs(r.Context(), req.Count, req.Service)

		case "tools":
			// seedToolUsage is disabled because we cannot access unexported metrics
			// seedToolUsage(req.Count, req.Service)
			logging.GetLogger().Warn("Seeding tools usage is currently disabled due to metrics encapsulation")

		case "health":
			seedHealthHistory(req.Service, req.Count)

		default:
			http.Error(w, "unknown seed type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"seeded"}`))
	}
}

func seedTraffic(tm *topology.Manager, _ string, count int) {
	points := make([]topology.TrafficPoint, 0, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		t := now.Add(time.Duration(-i) * time.Minute)
		points = append(points, topology.TrafficPoint{
			Time:    t.Format("15:04"),
			Total:   int64(100 + (i % 50)),
			Errors:  int64(i % 5),
			Latency: int64(20 + (i % 20)),
		})
	}
	tm.SeedTrafficHistory(points)
}

func seedLogs(ctx context.Context, count int, source string) {
	if source == "" {
		source = "seed-service"
	}
	for i := 0; i < count; i++ {
		level := "INFO"
		if i%10 == 0 {
			level = "ERROR"
		} else if i%5 == 0 {
			level = "WARN"
		}

		// Map string level to slog.Level manually since logging.ParseLevel doesn't exist
		var slogLevel slog.Level
		switch level {
		case "DEBUG":
			slogLevel = slog.LevelDebug
		case "INFO":
			slogLevel = slog.LevelInfo
		case "WARN":
			slogLevel = slog.LevelWarn
		case "ERROR":
			slogLevel = slog.LevelError
		default:
			slogLevel = slog.LevelInfo
		}

		logging.GetLogger().Log(ctx, slogLevel, fmt.Sprintf("Seeded log message %d", i), "source", source, "iteration", i)
	}
}

func seedHealthHistory(service string, count int) {
	for i := 0; i < count; i++ {
		status := "UP"
		if i%5 == 0 {
			status = "DOWN"
		}
		// Use AddHealthStatus instead of RecordCheck
		health.AddHealthStatus(service, status)
	}
}
