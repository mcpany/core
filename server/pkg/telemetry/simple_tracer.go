// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Span represents a single operation within a trace.
type Span struct {
	ID        string    `json:"id"`
	ParentID  string    `json:"parent_id,omitempty"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type spanCollector struct {
	mu    sync.Mutex
	spans []Span
}

type tracerContextKey struct{}
type parentSpanContextKey struct{}

var tracerKey = tracerContextKey{}
var parentSpanKey = parentSpanContextKey{}

// NewTraceContext initializes a new trace context.
func NewTraceContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, tracerKey, &spanCollector{
		spans: make([]Span, 0),
	})
}

// StartSpan starts a new span.
// It returns a new context with the span set as the active parent, and a function to end the span.
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	collector, ok := ctx.Value(tracerKey).(*spanCollector)
	if !ok {
		// No tracer in context, return no-op
		return ctx, func() {}
	}

	parentID, _ := ctx.Value(parentSpanKey).(string)

	var id string
	// Generate random ID
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback if random fails (unlikely)
		id = fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		id = hex.EncodeToString(bytes)
	}

	span := Span{
		ID:        id,
		ParentID:  parentID,
		Name:      name,
		StartTime: time.Now(),
	}

	// Return new context with this span as parent
	newCtx := context.WithValue(ctx, parentSpanKey, id)

	return newCtx, func() {
		span.EndTime = time.Now()
		collector.mu.Lock()
		collector.spans = append(collector.spans, span)
		collector.mu.Unlock()
	}
}

// GetSpans retrieves all collected spans from the context.
func GetSpans(ctx context.Context) []Span {
	collector, ok := ctx.Value(tracerKey).(*spanCollector)
	if !ok {
		return nil
	}
	collector.mu.Lock()
	defer collector.mu.Unlock()
	// Return copy
	result := make([]Span, len(collector.spans))
	copy(result, collector.spans)
	return result
}
