
package metrics

import (
	"io/ioutil"
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Check the response body for the expected metrics
	if !strings.Contains(string(body), "mcpany_test_counter 1") {
		t.Errorf("Expected metric mcpany_test_counter not found in response body")
	}
}
