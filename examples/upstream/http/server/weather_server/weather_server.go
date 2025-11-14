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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var weatherData = map[string]string{
	"new york": "Sunny, 25°C",
	"london":   "Cloudy, 15°C",
	"tokyo":    "Rainy, 20°C",
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	var location string
	if r.Method == "GET" {
		location = r.URL.Query().Get("location")
	} else if r.Method == "POST" {
		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		location = reqBody["location"]
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if location == "" {
		http.Error(w, "Missing location parameter", http.StatusBadRequest)
		return
	}

	weather, ok := weatherData[location]
	if !ok {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	response := map[string]string{
		"location": location,
		"weather":  weather,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var reqBody map[string]string
		if err := json.Unmarshal(msg, &reqBody); err != nil {
			log.Println(err)
			return
		}

		location := reqBody["location"]
		weather, ok := weatherData[location]
		if !ok {
			weather = "Location not found"
		}

		response := map[string]string{
			"location": location,
			"weather":  weather,
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8091"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/weather", weatherHandler)
	mux.HandleFunc("/ws", wsHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
