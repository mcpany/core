// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"sync"
	"time"
)

// Span represents a single operation in a trace.
type Span struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // e.g., "tool", "http", "middleware"
	StartTime    time.Time              `json:"startTime"`
	EndTime      time.Time              `json:"endTime"`
	Status       string                 `json:"status"` // "success", "error"
	Input        map[string]interface{} `json:"input,omitempty"`
	Output       map[string]interface{} `json:"output,omitempty"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	Children     []*Span                `json:"children,omitempty"`
	ServiceName  string                 `json:"serviceName,omitempty"`

	mu sync.Mutex
}

// AddChild adds a child span in a thread-safe way.
func (s *Span) AddChild(child *Span) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Children = append(s.Children, child)
}

// Collector defines the interface for collecting traces.
// (Keeping it for future use or if needed by Debugger implementation details)
type Collector interface {
	// AddSpan adds a completed span to the trace.
	AddSpan(traceID string, span *Span)
}
