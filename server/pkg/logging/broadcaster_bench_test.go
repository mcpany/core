// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
	"testing"
)

func BenchmarkBroadcaster_Broadcast(b *testing.B) {
	cases := []struct {
		name        string
		subscribers int
	}{
		{"1Sub", 1},
		{"10Subs", 10},
		{"100Subs", 100},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			broadcaster := NewBroadcaster()
			var wg sync.WaitGroup

			// Start subscribers
			for i := 0; i < tc.subscribers; i++ {
				ch := broadcaster.Subscribe()
				wg.Add(1)
				go func() {
					defer wg.Done()
					for range ch {
						// Drain channel
					}
				}()
			}

			msg := []byte("test log message")

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					broadcaster.Broadcast(msg)
				}
			})
			b.StopTimer()
		})
	}
}
