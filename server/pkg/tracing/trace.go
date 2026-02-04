// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tracing provides tracing functionality for the server.
package tracing

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Span represents a single span in a trace.
type Span struct {
	ID        string        `json:"id"`
	ParentID  string        `json:"parent_id,omitempty"`
	Name      string        `json:"name"`
	Type      string        `json:"type"` // "tool", "hook", "middleware"
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"` // "success", "error"
	Input     any           `json:"input,omitempty"`
	Output    any           `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`

	recorder *Recorder // internal reference
}

// Recorder records spans.
type Recorder struct {
	mu    sync.Mutex
	spans []*Span
}

// NewRecorder creates a new Recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		spans: make([]*Span, 0),
	}
}

// Add adds a span to the recorder.
func (r *Recorder) Add(s *Span) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, s)
}

// GetSpans returns all recorded spans.
func (r *Recorder) GetSpans() []*Span {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Return a copy to avoid race conditions if caller modifies slice
	spans := make([]*Span, len(r.spans))
	copy(spans, r.spans)
	return spans
}

type recorderKeyType struct{}
type activeSpanKeyType struct{}

var recorderKey = recorderKeyType{}
var activeSpanKey = activeSpanKeyType{}

// NewContext returns a new context with the given recorder.
func NewContext(ctx context.Context, r *Recorder) context.Context {
	return context.WithValue(ctx, recorderKey, r)
}

// StartSpan starts a new span.
// It checks the context for a parent span.
// It returns a new context containing the new span as the active span.
func StartSpan(ctx context.Context, name string, spanType string) (context.Context, *Span) {
	r, ok := ctx.Value(recorderKey).(*Recorder)
	if !ok || r == nil {
		// If no recorder, return a dummy span that won't be recorded
		return ctx, &Span{
			ID:        uuid.New().String(),
			Name:      name,
			Type:      spanType,
			StartTime: time.Now(),
			Status:    "pending",
		}
	}

	var parentID string
	if parentSpan, ok := ctx.Value(activeSpanKey).(*Span); ok {
		parentID = parentSpan.ID
	}

	span := &Span{
		ID:        uuid.New().String(),
		ParentID:  parentID,
		Name:      name,
		Type:      spanType,
		StartTime: time.Now(),
		Status:    "pending",
		recorder:  r,
	}

	// Update context with this span as active
	newCtx := context.WithValue(ctx, activeSpanKey, span)

	return newCtx, span
}

// End ends the span and records it.
func (s *Span) End() {
	if s.StartTime.IsZero() {
		return
	}
	s.Duration = time.Since(s.StartTime)
	if s.Status == "pending" {
		s.Status = "success"
	}
	if s.recorder != nil {
		s.recorder.Add(s)
	}
}

// SetError sets the error on the span.
func (s *Span) SetError(err error) {
	if err != nil {
		s.Error = err.Error()
		s.Status = "error"
	}
}

// SetInput sets the input on the span.
func (s *Span) SetInput(input any) {
	s.Input = input
}

// SetOutput sets the output on the span.
func (s *Span) SetOutput(output any) {
	s.Output = output
}
