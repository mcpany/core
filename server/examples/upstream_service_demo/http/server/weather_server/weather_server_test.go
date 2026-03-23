// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK\n", rr.Body.String())
}

func TestWeatherHandler_GET(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather?location=london", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "london", resp["location"])
	assert.Equal(t, "Cloudy, 15°C", resp["weather"])
}

func TestWeatherHandler_POST(t *testing.T) {
	reqBody := map[string]string{"location": "tokyo"}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/weather", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "tokyo", resp["location"])
	assert.Equal(t, "Rainy, 20°C", resp["weather"])
}

func TestWeatherHandler_POST_InvalidJSON(t *testing.T) {
	// Send invalid JSON
	req, err := http.NewRequest("POST", "/weather", strings.NewReader("{invalid_json"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request body")
}

func TestWeatherHandler_MissingLocation(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestWeatherHandler_LocationNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather?location=paris", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestWeatherHandler_InvalidMethod(t *testing.T) {
	req, err := http.NewRequest("PUT", "/weather", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestWSHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, dialResp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if dialResp != nil {
		defer func() { _ = dialResp.Body.Close() }()
	}
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	reqBody := map[string]string{"location": "new york"}
	err = conn.WriteJSON(reqBody)
	require.NoError(t, err)

	var resp map[string]string
	err = conn.ReadJSON(&resp)
	require.NoError(t, err)

	assert.Equal(t, "new york", resp["location"])
	assert.Equal(t, "Sunny, 25°C", resp["weather"])
}

func TestWSHandler_LocationNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, httpResp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = httpResp.Body.Close() }()
	defer func() { _ = conn.Close() }()

	reqBody := map[string]string{"location": "berlin"}
	err = conn.WriteJSON(reqBody)
	require.NoError(t, err)

	var resp map[string]string
	err = conn.ReadJSON(&resp)
	require.NoError(t, err)

	assert.Equal(t, "berlin", resp["location"])
	assert.Equal(t, "Location not found", resp["weather"])
}

func TestWSHandler_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, httpResp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = httpResp.Body.Close() }()
	defer func() { _ = conn.Close() }()

	// Send invalid JSON as text
	err = conn.WriteMessage(websocket.TextMessage, []byte("{invalid_json"))
	require.NoError(t, err)

	// The server should log error and close or stop responding.
	// Since the handler breaks the loop on error, the connection might close or just stop sending data.
	// We check if the connection is closed by trying to read from it.
	// We use a deadline to avoid hanging if the server doesn't close the connection.
	err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	require.NoError(t, err)

	_, _, err = conn.ReadMessage()
	require.Error(t, err)
}

func TestRun(t *testing.T) {
	// Use port 0 to let OS pick a free port
	args := []string{"weather-server", "-port", "0"}

	var addr string
	var runErr error

	// Retry loop for starting the server
	for i := 0; i < 20; i++ {
		stop := make(chan os.Signal, 1)
		ready := make(chan string, 1)
		errChan := make(chan error, 1)

		go func() {
			errChan <- run(args, stop, ready)
		}()

		select {
		case addr = <-ready:
			// Server started successfully
			// Run verification

			// Verify health check
			resp, err := http.Get("http://" + addr + "/health")
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Stop server
			stop <- syscall.SIGTERM

			// Wait for cleanup
			select {
			case err = <-errChan:
				assert.NoError(t, err)
			case <-time.After(2 * time.Second):
				t.Fatal("Server failed to stop in time")
			}
			return // Success!

		case runErr = <-errChan:
			// Server failed to start
			if strings.Contains(runErr.Error(), "address already in use") {
				t.Logf("Attempt %d: address in use, retrying... error=%v", i+1, runErr)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			t.Fatalf("Server failed to start: %v", runErr)

		case <-time.After(5 * time.Second):
			t.Logf("Attempt %d: timed out starting, retrying...", i+1)
			// Try to kill it if it's hanging?
			stop <- syscall.SIGKILL
			continue
		}
	}
	t.Fatalf("Server failed to start after 20 attempts: %v", runErr)
}
