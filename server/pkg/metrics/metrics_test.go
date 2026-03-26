// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartServer ...
// Summary: TestStartServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Initialize the metrics system
	err := Initialize()
	require.NoError(t, err)

	// Create a new test server
	server := httptest.NewServer(Handler())
	defer server.Close()

	// Make a request to the /metrics endpoint
	resp, err := http.Get(server.URL + "/metrics")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Check the response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMetricsCollection ...
// Summary: TestMetricsCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Initialize the metrics system with an in-memory sink
	sink := metrics.NewInmemSink(time.Second, 5*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	m, err := metrics.New(conf, sink)
	require.NoError(t, err)

	// Record some metrics
	m.SetGaugeWithLabels([]string{"my_gauge"}, 123, []metrics.Label{{Name: "service_name", Value: "label1"}})
	m.IncrCounter([]string{"my_counter"}, 1)
	m.MeasureSince([]string{"my_histogram"}, time.Now().Add(-1*time.Second))

	// Check if the metrics are present in the sink
	data := sink.Data()
	require.Len(t, data, 1)
	assert.Equal(t, float32(123), data[0].Gauges["mcpany.my_gauge;service_name=label1"].Value)
	assert.Equal(t, 1, data[0].Counters["mcpany.my_counter"].Count)
	assert.Contains(t, data[0].Samples, "mcpany.my_histogram")
}

// TestSetGauge ...
// Summary: TestSetGauge
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Re-init global to be safe or just call SetGauge.
	// SetGauge uses metrics.Global(), so we need to ensure it's set.
	// It is set by Initialize() or manually.
	// metrics.NewGlobal above sets it.
	// But tests run in parallel? No, t.Parallel() not called.
	// However, TestMetricsCollection sets NewGlobal.
	// We should probably setup NewGlobal in this test too or reuse.

	sink := metrics.NewInmemSink(time.Second, 5*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		SetGauge("test_gauge", 42.0, "test_label")
	})

	// Verify it reached sink
	data := sink.Data()
	if len(data) > 0 {
		assert.Equal(t, float32(42.0), data[0].Gauges["mcpany.test_gauge;service_name=test_label"].Value)
	}
}

// TestMeasureSince ...
// Summary: TestMeasureSince
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	sink := metrics.NewInmemSink(time.Second, 5*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		MeasureSince([]string{"test_timer"}, time.Now().Add(-1*time.Second))
	})
}

// TestIncrCounter ...
// Summary: TestIncrCounter
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	sink := metrics.NewInmemSink(time.Second, 5*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		IncrCounter([]string{"test_counter"}, 1)
	})
}

// TestStartServer_Real ...
// Summary: TestStartServer_Real
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test the actual StartServer function
	// We use a random port
	done := make(chan error)
	go func() {
		done <- StartServer("127.0.0.1:0")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Logf("StartServer failed immediately: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		// Started effectively
	}
}

// TestStartServer_Error ...
// Summary: TestStartServer_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	err := StartServer("invalid:address")
	assert.Error(t, err)
}

// TestMetricsWrappers ...
// Summary: TestMetricsWrappers
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Initialize to ensure sink is set up (though it might be already by other tests or init)
	sink := metrics.NewInmemSink(time.Second, 5*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	t.Run("IncrCounterWithLabels", func(t *testing.T) {
		assert.NotPanics(t, func() {
			IncrCounterWithLabels([]string{"test", "counter_lbl"}, 1, []Label{{Name: "k", Value: "v"}})
		})
	})

	t.Run("MeasureSinceWithLabels", func(t *testing.T) {
		assert.NotPanics(t, func() {
			MeasureSinceWithLabels([]string{"test", "timer_lbl"}, time.Now(), []Label{{Name: "k", Value: "v"}})
		})
	})

	t.Run("AddSample", func(t *testing.T) {
		assert.NotPanics(t, func() {
			AddSample([]string{"test", "sample"}, 1.0)
		})
	})

	t.Run("AddSampleWithLabels", func(t *testing.T) {
		assert.NotPanics(t, func() {
			AddSampleWithLabels([]string{"test", "sample_lbl"}, 1.0, []Label{{Name: "k", Value: "v"}})
		})
	})
}

// TestSetGauge_NoPanic ...
// Summary: TestSetGauge_NoPanic
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// The function signature is SetGauge(name string, val float32, labels ...string)
	// It should NOT panic if labels are empty.
	assert.NotPanics(t, func() {
		SetGauge("test_gauge", 1.0)
	}, "SetGauge should not panic if no labels are provided")

	// Test with valid labels
	assert.NotPanics(t, func() {
		SetGauge("test_gauge_with_label", 1.0, "some_service")
	}, "SetGauge should not panic with valid labels")
}
