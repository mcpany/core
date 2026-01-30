// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tracing provides functionality for handling and exporting traces.
package tracing

import (
	"context"
	"encoding/json"
	"sort"
	"sync"

	"go.opentelemetry.io/otel/sdk/trace"
)

// Span represents a span in a trace.
type Span struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	StartTime    int64          `json:"startTime"` // Unix millis
	EndTime      int64          `json:"endTime"`   // Unix millis
	Status       string         `json:"status"`    // success, error, pending
	Input        map[string]any `json:"input,omitempty"`
	Output       map[string]any `json:"output,omitempty"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
	Children     []*Span        `json:"children,omitempty"`
	ParentID     string         `json:"-"` // Internal use for reconstruction
}

// Trace represents a full trace.
type Trace struct {
	ID            string `json:"id"`
	RootSpan      *Span  `json:"rootSpan"`
	Timestamp     string `json:"timestamp"` // ISO 8601
	TotalDuration int64  `json:"totalDuration"`
	Status        string `json:"status"`
	Trigger       string `json:"trigger"`
}

// InMemoryExporter stores spans in memory.
type InMemoryExporter struct {
	mu    sync.RWMutex
	spans []trace.ReadOnlySpan
	limit int

	// Separate storage for seeded traces which might not come from OTEL spans
	seededTraces []*Trace
}

// NewInMemoryExporter creates a new InMemoryExporter.
func NewInMemoryExporter(limit int) *InMemoryExporter {
	return &InMemoryExporter{
		limit: limit,
	}
}

// ExportSpans exports a batch of spans.
func (e *InMemoryExporter) ExportSpans(_ context.Context, spans []trace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Append new spans
	e.spans = append(e.spans, spans...)

	// Trim if needed
	if len(e.spans) > e.limit {
		e.spans = e.spans[len(e.spans)-e.limit:]
	}

	return nil
}

// Shutdown shuts down the exporter.
func (e *InMemoryExporter) Shutdown(_ context.Context) error {
	return nil
}

// Seed adds manual traces to the store.
func (e *InMemoryExporter) Seed(traces []*Trace) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.seededTraces = append(e.seededTraces, traces...)
}

// GetTraces returns reconstructed traces from stored spans and seeded traces.
func (e *InMemoryExporter) GetTraces() []*Trace {
	e.mu.RLock()
	defer e.mu.RUnlock()

	allTraces := make([]*Trace, 0)

	// 1. Add seeded traces
	allTraces = append(allTraces, e.seededTraces...)

	// 2. Reconstruct traces from OTEL spans
	// Group spans by TraceID
	spansByTraceID := make(map[string][]trace.ReadOnlySpan)
	for _, s := range e.spans {
		if !s.SpanContext().HasTraceID() {
			continue
		}
		tid := s.SpanContext().TraceID().String()
		spansByTraceID[tid] = append(spansByTraceID[tid], s)
	}

	for tid, spans := range spansByTraceID {
		// Convert ReadOnlySpan to our Span struct
		var spanMap = make(map[string]*Span)
		var root *Span
		// var minStart int64 = -1
		// var maxEnd int64 = -1

		// First pass: Create all spans
		for _, s := range spans {
			start := s.StartTime().UnixMilli()
			end := s.EndTime().UnixMilli()

			// Attributes to Input/Output/Type mapping
			input := make(map[string]any)
			output := make(map[string]any)

			for _, attr := range s.Attributes() {
				key := string(attr.Key)
				// attr.Value.AsString() is safe
				val := attr.Value.AsString()
				// Basic heuristic
				switch {
				case key == "input" || key == "arguments":
					_ = json.Unmarshal([]byte(val), &input)
				case key == "output" || key == "result":
					_ = json.Unmarshal([]byte(val), &output)
				default:
					input[key] = val
				}
			}

			status := "success"
			if s.Status().Code == 2 { // Error
				status = "error"
			}

			span := &Span{
				ID:        s.SpanContext().SpanID().String(),
				Name:      s.Name(),
				Type:      "tool", // Default
				StartTime: start,
				EndTime:   end,
				Status:    status,
				Input:     input,
				Output:    output,
				ErrorMessage: s.Status().Description,
			}

			if s.Parent().IsValid() {
				span.ParentID = s.Parent().SpanID().String()
			}

			spanMap[span.ID] = span
		}

		// Second pass: Build hierarchy
		for _, span := range spanMap {
			if span.ParentID != "" {
				parent, ok := spanMap[span.ParentID]
				if ok {
					parent.Children = append(parent.Children, span)
				} else if root == nil {
					// Fallback: If parent not found (orphaned in this view), make it root
					root = span
				}
			} else {
				root = span
			}
		}

		if root != nil {
			t := &Trace{
				ID:            tid,
				RootSpan:      root,
				Timestamp:     "", // TODO: Format time
				TotalDuration: 0,  // TODO: Calc duration
				Status:        root.Status,
				Trigger:       "user",
			}
			allTraces = append(allTraces, t)
		}
	}

	// Sort traces by timestamp descending
	sort.Slice(allTraces, func(i, j int) bool {
		return allTraces[i].Timestamp > allTraces[j].Timestamp
	})

	return allTraces
}
