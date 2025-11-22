// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may
// obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestWeatherHandler(t *testing.T) {
	// Test GET request
	req, err := http.NewRequest("GET", "/weather?location=london", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"location":"london","weather":"Cloudy, 15°C"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Test POST request
	req, err = http.NewRequest("POST", "/weather",
		bytes.NewBuffer([]byte(`{"location": "tokyo"}`)))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected = `{"location":"tokyo","weather":"Rainy, 20°C"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Test missing location
	req, err = http.NewRequest("GET", "/weather", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestWsHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer s.Close()

	wsURL := "ws" + s.URL[4:]
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Test single message
	err = ws.WriteJSON(map[string]string{"location": "new york"})
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	var resp map[string]string
	err = ws.ReadJSON(&resp)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	expected := map[string]string{"location": "new york", "weather": "Sunny, 25°C"}
	if resp["location"] != expected["location"] || resp["weather"] != expected["weather"] {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp, expected)
	}

	// Test multiple messages
	for i := 0; i < 3; i++ {
		err = ws.WriteJSON(map[string]string{"location": "london"})
		if err != nil {
			t.Fatalf("Failed to write message: %v", err)
		}

		err = ws.ReadJSON(&resp)
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		expected = map[string]string{"location": "london", "weather": "Cloudy, 15°C"}
		if resp["location"] != expected["location"] || resp["weather"] != expected["weather"] {
			t.Errorf("handler returned unexpected body: got %v want %v",
				resp, expected)
		}
	}
}
