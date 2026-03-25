// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"
	"time"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	err := Initialize()
	assert.NoError(t, err)
}

func TestFunctionalMetrics(t *testing.T) {
	sink := armonmetrics.NewInmemSink(time.Hour, time.Hour)
	conf := armonmetrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := armonmetrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	t.Run("Counter", func(t *testing.T) {
		IncrCounter([]string{"test", "counter"}, 1)
		IncrCounterWithLabels([]string{"test", "labeled_counter"}, 5, []Label{{Name: "foo", Value: "bar"}})

		data := sink.Data()
		require.NotEmpty(t, data)

		foundBase := false
		foundLabeled := false
		for _, v := range data {
			for _, c := range v.Counters {
				if c.Name == "mcpany.test.counter" && c.Sum >= 1 {
					foundBase = true
				}
				if c.Name == "mcpany.test.labeled_counter" && c.Sum >= 5 {
					for _, l := range c.Labels {
						if l.Name == "foo" && l.Value == "bar" {
							foundLabeled = true
						}
					}
				}
			}
		}
		assert.True(t, foundBase, "Base counter not found")
		assert.True(t, foundLabeled, "Labeled counter not found")
	})

	t.Run("Gauge", func(t *testing.T) {
		SetGauge("test_gauge", 42, "my-service")

		data := sink.Data()
		require.NotEmpty(t, data)
		found := false
		for _, v := range data {
			for _, g := range v.Gauges {
				if g.Name == "mcpany.test_gauge" && g.Value == 42 {
					for _, l := range g.Labels {
						if l.Name == "service_name" && l.Value == "my-service" {
							found = true
						}
					}
				}
			}
		}
		assert.True(t, found, "Gauge with service label not found")
	})

	t.Run("Samples and Latency", func(t *testing.T) {
		MeasureSince([]string{"test", "latency"}, time.Now().Add(-100*time.Millisecond))
		AddSample([]string{"test", "sample"}, 10.5)

		data := sink.Data()
		require.NotEmpty(t, data)

		foundLatency := false
		foundSample := false
		for _, v := range data {
			for _, s := range v.Samples {
				if s.Name == "mcpany.test.latency" {
					foundLatency = true
				}
				if s.Name == "mcpany.test.sample" && s.Sum >= 10.5 {
					foundSample = true
				}
			}
		}
		assert.True(t, foundLatency, "Latency sample not found")
		assert.True(t, foundSample, "Data sample not found")
	})
}
