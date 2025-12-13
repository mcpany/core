/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func TestStartServer(t *testing.T) {
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

func TestMetricsCollection(t *testing.T) {
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
