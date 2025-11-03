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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	Initialize()
	os.Exit(m.Run())
}

func TestMetrics(t *testing.T) {
	// Test Handler
	if handler := Handler(); handler == nil {
		t.Error("Handler returned a nil handler")
	}

	// Test SetGauge
	SetGauge("test_gauge", 1.0, "test_service")

	// Test IncrCounter
	IncrCounter([]string{"test_counter"}, 1.0)

	// Test MeasureSince
	MeasureSince([]string{"test_measurement"}, time.Now())

	// Create a test server
	ts := httptest.NewServer(Handler())
	defer ts.Close()

	// Make a request to the /metrics endpoint
	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Check the response body for the expected metrics
	if !strings.Contains(string(body), "mcpany_test_counter 1") {
		t.Errorf("Expected metric mcpany_test_counter not found in response body")
	}
}
