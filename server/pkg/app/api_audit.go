// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pdf/fpdf"
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

	format := r.URL.Query().Get("format")
	if format == "pdf" {
		a.exportAuditPDF(w, entries)
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

func (a *Application) exportAuditPDF(w http.ResponseWriter, entries []audit.Entry) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetFont("Arial", "B", 16)
	pdf.AddPage()
	pdf.Cell(40, 10, "Audit Log Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 10)
	// Headers
	headers := []string{"Timestamp", "ToolName", "User", "Status", "Duration"}
	widths := []float64{50, 60, 40, 30, 30}

	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
	for _, entry := range entries {
		status := "Success"
		if entry.Error != "" {
			status = "Error"
		}
		pdf.CellFormat(widths[0], 6, entry.Timestamp.Format("2006-01-02 15:04:05"), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[1], 6, entry.ToolName, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[2], 6, entry.UserID, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[3], 6, status, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[4], 6, fmt.Sprintf("%d ms", entry.DurationMs), "1", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit_export_%s.pdf", time.Now().Format("20060102_150405")))

	if err := pdf.Output(w); err != nil {
		// Log error if possible
	}
}
