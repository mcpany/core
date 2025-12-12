// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

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
	defer httpResp.Body.Close()
	defer conn.Close()

	reqBody := map[string]string{"location": "berlin"}
	err = conn.WriteJSON(reqBody)
	require.NoError(t, err)

	var resp map[string]string
	err = conn.ReadJSON(&resp)
	require.NoError(t, err)

	assert.Equal(t, "berlin", resp["location"])
	assert.Equal(t, "Location not found", resp["weather"])
}
