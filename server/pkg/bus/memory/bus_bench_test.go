// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"testing"
)

func BenchmarkPublish(b *testing.B) {
	bus := New[string]()
	ctx := context.Background()
	topic := "bench-topic"

	// Add subscribers
	for i := 0; i < 10; i++ {
		bus.Subscribe(ctx, topic, func(msg string) {
			// consume
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = bus.Publish(ctx, topic, "test message")
	}
}
