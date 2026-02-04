// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"context"
	"errors"
	"testing"
)

func TestTracing(t *testing.T) {
	ctx := context.Background()
	recorder := NewRecorder()
	ctx = NewContext(ctx, recorder)

	// Start root span
	ctx, rootSpan := StartSpan(ctx, "root", "test")
	if rootSpan == nil {
		t.Fatal("rootSpan is nil")
	}
	if rootSpan.ID == "" {
		t.Error("rootSpan ID is empty")
	}

	// Start child span
	_, childSpan := StartSpan(ctx, "child", "test")
	if childSpan == nil {
		t.Fatal("childSpan is nil")
	}
	if childSpan.ParentID != rootSpan.ID {
		t.Errorf("expected parent ID %s, got %s", rootSpan.ID, childSpan.ParentID)
	}

	childSpan.SetInput("input")
	childSpan.SetOutput("output")
	childSpan.End()

	rootSpan.SetError(errors.New("test error"))
	rootSpan.End()

	spans := recorder.GetSpans()
	if len(spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(spans))
	}

	// Verify order (usually child ends first, but recorder order depends on End() call)
	// We called childSpan.End() first.
	if spans[0].Name != "child" {
		t.Errorf("expected first span to be 'child', got %s", spans[0].Name)
	}
	if spans[1].Name != "root" {
		t.Errorf("expected second span to be 'root', got %s", spans[1].Name)
	}

	if spans[0].Input != "input" {
		t.Errorf("expected input 'input', got %v", spans[0].Input)
	}
	if spans[1].Error != "test error" {
		t.Errorf("expected error 'test error', got %v", spans[1].Error)
	}
}

func TestNoRecorder(t *testing.T) {
	ctx := context.Background()
	// No recorder in context
	_, span := StartSpan(ctx, "test", "test")
	if span == nil {
		t.Fatal("span should not be nil even without recorder")
	}
	// Should not panic
	span.End()
}
