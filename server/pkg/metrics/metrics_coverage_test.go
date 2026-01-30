// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGauge_Labels(t *testing.T) {
	// Setup in-memory sink to inspect metrics
	sink := metrics.NewInmemSink(10*time.Second, 10*time.Minute)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	// Test case 1: Single label
	// Expected behavior: SetGauge("name", val, "foo") -> service_name="foo"
	SetGauge("gauge_single", 1.0, "foo")

	// Flush to ensure metrics are in sink (InmemSink might need time or manual flush?
	// InmemSink.Data() returns current data).
	data := sink.Data()
	require.NotEmpty(t, data)

	// The key format for go-metrics with labels depends on the sink, but InmemSink usually flattens them
	// or supports structured access. InmemSink.Data() returns []Sample.
	// However, go-metrics InmemSink keys are usually "name;label=val;...".

	found := false
	for _, s := range data {
		if val, ok := s.Gauges["mcpany.gauge_single;service_name=foo"]; ok {
			assert.Equal(t, float32(1.0), val.Value)
			found = true
			break
		}
	}
	assert.True(t, found, "Expected gauge_single with service_name=foo")

	// Test case 2: Multiple labels (Confusing API behavior documentation)
	// Expected behavior: SetGauge("name", val, "foo", "bar") -> service_name="foo" (bar ignored)
	SetGauge("gauge_multi", 2.0, "foo", "bar")

	data = sink.Data()
	found = false
	for _, s := range data {
		if val, ok := s.Gauges["mcpany.gauge_multi;service_name=foo"]; ok {
			assert.Equal(t, float32(2.0), val.Value)
			found = true
			break
		}
	}
	assert.True(t, found, "Expected gauge_multi with service_name=foo, ignoring extra args")
}

func TestWrappers(t *testing.T) {
	// Ensure these wrappers don't panic and pass data correctly
	sink := metrics.NewInmemSink(10*time.Second, 10*time.Minute)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		IncrCounter([]string{"wrapper", "counter"}, 1)
		MeasureSince([]string{"wrapper", "timer"}, time.Now())
		AddSample([]string{"wrapper", "sample"}, 10)
	})

	data := sink.Data()
	require.NotEmpty(t, data)

	// Verify Counter
	foundCounter := false
	for _, s := range data {
		if c, ok := s.Counters["mcpany.wrapper.counter"]; ok {
			assert.Equal(t, 1, c.Count) // InmemSink stores Count as int? struct { Count int; ... }
			foundCounter = true
			break
		}
	}
	assert.True(t, foundCounter, "Counter wrapper failed")
}
