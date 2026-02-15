package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK\n", rr.Body.String())
}

func TestTimeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/time", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(timeHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "current_time")
	assert.Contains(t, response, "timezone")
	assert.Equal(t, "UTC", response["timezone"])

	// Parse the time to make sure it's valid RFC3339
	parsedTime, err := time.Parse(time.RFC3339, response["current_time"])
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), parsedTime, 2*time.Second)
}
