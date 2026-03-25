// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"
	"time"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestMetricsCoverage(t *testing.T) {
	sink := armonmetrics.NewInmemSink(10*time.Second, 10*time.Minute)
	conf := armonmetrics.DefaultConfig("mcpany")
	_, err := armonmetrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	IncrCounter([]string{"test", "counter"}, 1)
	IncrCounterWithLabels([]string{"test", "counter", "labels"}, 1, []Label{{Name: "l1", Value: "v1"}})
	SetGauge("test_gauge", 42)
	MeasureSince([]string{"test", "latency"}, time.Now().Add(-time.Second))
	MeasureSinceWithLabels([]string{"test", "latency", "labels"}, time.Now().Add(-time.Second), []Label{{Name: "l1", Value: "v1"}})
	AddSample([]string{"test", "sample"}, 0.5)
	AddSampleWithLabels([]string{"test", "sample", "labels"}, 0.5, []Label{{Name: "l1", Value: "v1"}})
}
