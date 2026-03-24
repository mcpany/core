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

func TestSetGauge(t *testing.T) {
	sink := armonmetrics.NewInmemSink(time.Second, 5*time.Second)
	conf := armonmetrics.DefaultConfig("mcpany")
	m, err := armonmetrics.New(conf, sink)
	require.NoError(t, err)
	m.SetGaugeWithLabels([]string{"my_gauge"}, 123, []armonmetrics.Label{{Name: "service_name", Value: "label1"}})
}

func TestIncrCounter(t *testing.T) {
	sink := armonmetrics.NewInmemSink(time.Second, 5*time.Second)
	conf := armonmetrics.DefaultConfig("mcpany")
	_, err := armonmetrics.NewGlobal(conf, sink)
	require.NoError(t, err)
	IncrCounter([]string{"test_counter"}, 1)
}

func TestIncrCounterWithLabels(t *testing.T) {
	sink := armonmetrics.NewInmemSink(time.Second, 5*time.Second)
	conf := armonmetrics.DefaultConfig("mcpany")
	_, err := armonmetrics.NewGlobal(conf, sink)
	require.NoError(t, err)
	IncrCounterWithLabels([]string{"test_counter_labels"}, 1, []Label{{Name: "label1", Value: "val1"}})
}

func TestMeasureSince(t *testing.T) {
	sink := armonmetrics.NewInmemSink(time.Second, 5*time.Second)
	conf := armonmetrics.DefaultConfig("mcpany")
	_, err := armonmetrics.NewGlobal(conf, sink)
	require.NoError(t, err)
	MeasureSince([]string{"test_latency"}, time.Now().Add(-100*time.Millisecond))
}

func TestMeasureSinceWithLabels(t *testing.T) {
	sink := armonmetrics.NewInmemSink(time.Second, 5*time.Second)
	conf := armonmetrics.DefaultConfig("mcpany")
	_, err := armonmetrics.NewGlobal(conf, sink)
	require.NoError(t, err)
	MeasureSinceWithLabels([]string{"test_latency_labels"}, time.Now().Add(-100*time.Millisecond), []Label{{Name: "label1", Value: "val1"}})
}
