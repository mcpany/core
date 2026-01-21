/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// AuditReader defines the interface for reading audit logs.
type AuditReader interface {
	Read(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
}

// ExportCSV exports audit logs to CSV format.
func ExportCSV(ctx context.Context, reader AuditReader, filter AuditFilter, w io.Writer) error {
	entries, err := reader.Read(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to read audit entries: %w", err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Timestamp",
		"Tool Name",
		"User ID",
		"Profile ID",
		"Duration (ms)",
		"Error",
		"Arguments",
		"Result",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, entry := range entries {
		record := []string{
			entry.Timestamp.Format(time.RFC3339),
			entry.ToolName,
			entry.UserID,
			entry.ProfileID,
			fmt.Sprintf("%d", entry.DurationMs),
			entry.Error,
			string(entry.Arguments),
			fmt.Sprintf("%v", entry.Result),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}
