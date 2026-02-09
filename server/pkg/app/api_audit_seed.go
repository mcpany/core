// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
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

	// Default count
	count := 50
	if c := r.URL.Query().Get("count"); c != "" {
		if val, err := strconv.Atoi(c); err == nil {
			count = val
		}
	} else {
		// Try JSON body
		var body struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.Count > 0 {
			count = body.Count
		}
	}

	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit middleware not initialized", http.StatusServiceUnavailable)
		return
	}

	// Fake data generators
	users := []string{"admin", "alice", "bob", "ci-bot", "developer"}
	tools := []string{
		"weather_get", "github_issue_create", "github_pr_list",
		"postgres_query", "slack_post_message", "filesystem_read_file",
		"brave_search", "memory_store",
	}
	profiles := []string{"default", "prod", "staging"}

	generated := 0
	for i := 0; i < count; i++ {
		// Random time within last 24 hours
		randomDuration := time.Duration(rand.Int63n(int64(24 * time.Hour)))
		timestamp := time.Now().Add(-randomDuration)

		toolName := tools[rand.Intn(len(tools))]
		user := users[rand.Intn(len(users))]
		profile := profiles[rand.Intn(len(profiles))]

		// Random success/failure
		var errStr string
		if rand.Float32() < 0.1 { // 10% failure rate
			errStr = "simulated error: connection timeout"
		}

		duration := time.Duration(rand.Intn(5000)) * time.Millisecond

		args := json.RawMessage(fmt.Sprintf(`{"q": "query_%d"}`, i))
		result := map[string]string{"status": "ok"}

		entry := audit.Entry{
			Timestamp:  timestamp,
			ToolName:   toolName,
			UserID:     user,
			ProfileID:  profile,
			Arguments:  args,
			Result:     result,
			Error:      errStr,
			Duration:   duration.String(),
			DurationMs: duration.Milliseconds(),
		}

		if err := a.standardMiddlewares.Audit.Write(r.Context(), entry); err != nil {
			// Log error but continue
			fmt.Printf("Failed to seed audit log: %v\n", err)
		} else {
			generated++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"generated": generated,
		"requested": count,
	})
}
