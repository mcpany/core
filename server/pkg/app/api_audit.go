// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func (a *Application) handleAuditExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Basic filtering from query params
	filter := audit.Filter{}
	if start := r.URL.Query().Get("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = &t
		}
	}
	if end := r.URL.Query().Get("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndTime = &t
		}
	}
	filter.ToolName = r.URL.Query().Get("tool_name")
	filter.UserID = r.URL.Query().Get("user_id")

	// Get the audit store from standard middlewares
	// Note: We need to ensure standardMiddlewares is accessible.
	// Based on server.go discovery, it is a field on Application.
	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit store not configured", http.StatusServiceUnavailable)
		return
	}

	entries, err := a.standardMiddlewares.Audit.Read(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to read audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit_export_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	_ = writer.Write([]string{"Timestamp", "ToolName", "UserID", "ProfileID", "Arguments", "Result", "Error", "DurationMs"})

	for _, entry := range entries {
		_ = writer.Write([]string{
			entry.Timestamp.Format(time.RFC3339Nano),
			entry.ToolName,
			entry.UserID,
			entry.ProfileID,
			string(entry.Arguments),
			fmt.Sprintf("%v", entry.Result),
			entry.Error,
			fmt.Sprintf("%d", entry.DurationMs),
		})
	}
}

func (a *Application) handleDebugSeedAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Count int `json:"count"`
	}
	// Default count
	req.Count = 50

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		// ignore error if body is empty, use default
	}

	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit middleware not configured", http.StatusServiceUnavailable)
		return
	}

	users := []string{"alice", "bob", "charlie", "system-admin", "dev-ops"}
	tools := []string{"weather_get", "github_issue_create", "github_pr_list", "postgres_query", "slack_post_message", "sentry_list_issues"}
	profiles := []string{"default", "prod", "staging"}

	now := time.Now()
	rng := rand.New(rand.NewSource(now.UnixNano()))

	generated := 0
	for i := 0; i < req.Count; i++ {
		// Random time within last 24h
		offset := time.Duration(rng.Int63n(int64(24 * time.Hour)))
		timestamp := now.Add(-offset)

		user := users[rng.Intn(len(users))]
		toolName := tools[rng.Intn(len(tools))]
		profileID := profiles[rng.Intn(len(profiles))]

		isError := rng.Intn(10) < 2 // 20% error rate
		var errStr string
		if isError {
			errStr = "simulated error: connection timeout"
		}

		duration := time.Duration(rng.Int63n(int64(2 * time.Second)))

		entry := audit.Entry{
			Timestamp:  timestamp,
			ToolName:   toolName,
			UserID:     user,
			ProfileID:  profileID,
			Arguments:  json.RawMessage(`{"simulated": true}`),
			Result:     map[string]interface{}{"status": "simulated_success"},
			Error:      errStr,
			Duration:   duration.String(),
			DurationMs: duration.Milliseconds(),
		}

		if err := a.standardMiddlewares.Audit.Write(r.Context(), entry); err != nil {
			// Log but continue
			fmt.Printf("Failed to seed log: %v\n", err)
		} else {
			generated++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"generated": generated,
	})
}
