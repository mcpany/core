// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func (a *Application) handleDebugSeedAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse count from query param or body
	count := 50
	if c := r.URL.Query().Get("count"); c != "" {
		if val, err := strconv.Atoi(c); err == nil {
			count = val
		}
	} else {
		// Try JSON body
		var req struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Count > 0 {
			count = req.Count
		}
	}

	// Limit count
	if count > 1000 {
		count = 1000
	}

	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit middleware not initialized", http.StatusServiceUnavailable)
		return
	}

	// Tools list
	tools := []string{
		"weather_get", "github_issue_create", "postgres_query", "slack_post_message",
		"google_calendar_list_events", "linear_issue_update", "sentry_list_projects",
		"memory_store", "filesystem_read_file", "brave_search",
	}

	// Users list
	users := []string{"alice", "bob", "charlie", "david", "eve", "frank", "grace", "system-admin"}

	// Profiles list
	profiles := []string{"default", "prod", "staging", "dev"}

	// Generate logs
	//nolint:gosec // Seeding random data for debug
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	generated := 0
	for i := 0; i < count; i++ {
		// Random time within last 24 hours
		offset := time.Duration(rng.Int63n(int64(24 * time.Hour)))
		ts := now.Add(-offset)

		toolName := tools[rng.Intn(len(tools))]
		userID := users[rng.Intn(len(users))]
		profileID := profiles[rng.Intn(len(profiles))]
		durationMs := int64(rng.Intn(500) + 10) // 10-510ms

		entry := audit.Entry{
			Timestamp:  ts,
			ToolName:   toolName,
			UserID:     userID,
			ProfileID:  profileID,
			DurationMs: durationMs,
			Duration:   fmt.Sprintf("%dms", durationMs),
		}

		// Random args/result
		args := map[string]any{
			"query": "something",
			"id":    rng.Intn(1000),
			"force": rng.Intn(2) == 1,
		}
		argsBytes, _ := json.Marshal(args)
		entry.Arguments = json.RawMessage(argsBytes)

		// 10% chance of error
		if rng.Float32() < 0.1 {
			entry.Error = "simulated error: connection timeout"
			entry.Result = nil
		} else {
			entry.Result = map[string]any{"status": "ok", "data": "some data"}
		}

		if err := a.standardMiddlewares.Audit.Write(context.Background(), entry); err != nil {
			// Log error but continue
			fmt.Printf("Failed to seed audit log: %v\n", err)
		} else {
			generated++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":   fmt.Sprintf("Seeded %d audit logs", generated),
		"generated": generated,
	})
}
