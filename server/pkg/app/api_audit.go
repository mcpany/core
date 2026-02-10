// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
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
	if a.standardMiddlewares.Audit == nil {
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

	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit store not configured", http.StatusServiceUnavailable)
		return
	}

	// Parse count
	count := 50
	if cStr := r.URL.Query().Get("count"); cStr != "" {
		if c, err := strconv.Atoi(cStr); err == nil && c > 0 {
			count = c
		}
	}

	// Generate fake data
	users := []string{"alice", "bob", "charlie", "dave", "eve", "frank", "system-admin"}
	tools := []string{
		"weather_get", "github_issue_create", "github_pr_list",
		"postgres_query", "slack_send_message", "memory_store",
		"calculator_add", "web_search", "image_generate",
	}
	profiles := []string{"default", "prod", "staging", "dev"}

	// Generate logs across the last 24 hours
	now := time.Now()
	seeded := 0

	for i := 0; i < count; i++ {
		// Random time within last 24h
		offset := time.Duration(rand.Int63n(int64(24 * time.Hour))) //nolint:gosec // Usage for debug seeding
		ts := now.Add(-offset)

		toolName := tools[rand.Intn(len(tools))]       //nolint:gosec // Usage for debug seeding
		user := users[rand.Intn(len(users))]           //nolint:gosec // Usage for debug seeding
		profile := profiles[rand.Intn(len(profiles))] //nolint:gosec // Usage for debug seeding

		success := rand.Float32() > 0.1 //nolint:gosec // Usage for debug seeding

		var errStr string
		var result any

		if success {
			result = map[string]string{"status": "ok", "data": "simulated_success"}
		} else {
			if rand.Float32() > 0.5 { //nolint:gosec // Usage for debug seeding
				errStr = "connection timeout"
			} else {
				errStr = "invalid arguments"
			}
		}

		durationMs := rand.Int63n(2000) //nolint:gosec // Usage for debug seeding
		duration := time.Duration(durationMs) * time.Millisecond

		entry := audit.Entry{
			Timestamp:  ts,
			ToolName:   toolName,
			UserID:     user,
			ProfileID:  profile,
			Arguments:  json.RawMessage(`{"arg": "simulated"}`),
			Result:     result,
			Error:      errStr,
			Duration:   duration.String(),
			DurationMs: durationMs,
		}

		if err := a.standardMiddlewares.Audit.Write(r.Context(), entry); err == nil {
			seeded++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"seeded":  seeded,
		"message": fmt.Sprintf("Seeded %d audit logs", seeded),
	})
}
