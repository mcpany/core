// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package trace provides a lightweight tracing system for the application.
package trace

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type traceContextKey struct{}

// FromContext returns the current parent span from the context.
func FromContext(ctx context.Context) *Span {
	if val, ok := ctx.Value(traceContextKey{}).(*Span); ok {
		return val
	}
	return nil
}

// NewContext returns a new context with the given span as the parent.
func NewContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, traceContextKey{}, span)
}

// StartSpan starts a new span.
// If a parent span is found in the context, the new span is added as a child.
func StartSpan(ctx context.Context, name string, spanType string) (context.Context, *Span) {
	parent := FromContext(ctx)

	span := &Span{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      spanType,
		StartTime: time.Now(),
		Status:    "success", // Default
		Children:  make([]*Span, 0),
	}

	if parent != nil {
		parent.AddChild(span)
	}

	return NewContext(ctx, span), span
}

// EndSpan ends the span.
func EndSpan(span *Span) {
	span.EndTime = time.Now()
}
