// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

// handleDebugSeedAuditLogs seeds the audit logs with fake data for testing/demo purposes.
// It generates `count` (default 50) entries.
func (a *Application) handleDebugSeedAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Count int `json:"count"`
	}
	// Try to decode, ignore error and use default if empty
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Count <= 0 {
		req.Count = 50
	}
	if req.Count > 1000 {
		req.Count = 1000 // Limit
	}

	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit middleware not initialized", http.StatusServiceUnavailable)
		return
	}

	users := []string{"alice", "bob", "charlie", "dave", "eve", "system-admin"}
	tools := []string{
		"weather_get", "weather_forecast",
		"github_issue_create", "github_pr_list", "github_repo_read",
		"postgres_query", "postgres_schema",
		"slack_message_send", "slack_channel_list",
		"brave_search",
		"filesystem_read_file", "filesystem_list_dir",
	}
	profiles := []string{"default", "prod", "dev", "staging"}

	now := time.Now()
	generated := 0

	for i := 0; i < req.Count; i++ {
		// Random time within last 24 hours
		offset := time.Duration(rand.Int63n(int64(24 * time.Hour)))
		ts := now.Add(-offset)

		toolName := tools[rand.Intn(len(tools))]
		user := users[rand.Intn(len(users))]
		profile := profiles[rand.Intn(len(profiles))]
		durationMs := rand.Int63n(2000) + 10 // 10ms to 2000ms

		entry := audit.Entry{
			Timestamp:  ts,
			ToolName:   toolName,
			UserID:     user,
			ProfileID:  profile,
			DurationMs: durationMs,
			Duration:   fmt.Sprintf("%dms", durationMs),
		}

		// Random arguments
		args := map[string]interface{}{
			"query": "something",
			"limit": rand.Intn(100),
		}
		argsBytes, _ := json.Marshal(args)
		entry.Arguments = json.RawMessage(argsBytes)

		// Random success/failure (90% success)
		if rand.Float32() > 0.9 {
			entry.Error = "simulated error: connection timeout or invalid input"
		} else {
			res := map[string]interface{}{
				"status": "ok",
				"data":   "some result data",
			}
			entry.Result = res
		}

		if err := a.standardMiddlewares.Audit.Write(r.Context(), entry); err != nil {
			// Log but continue
			fmt.Printf("Failed to seed log: %v\n", err)
		} else {
			generated++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"generated": generated,
		"message":   fmt.Sprintf("Seeded %d audit logs", generated),
	})
}
