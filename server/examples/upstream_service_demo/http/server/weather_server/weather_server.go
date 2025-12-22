// Package main implements the weather server for the upstream service demo.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

var weatherData = map[string]string{
	"new york": "Sunny, 25°C",
	"london":   "Cloudy, 15°C",
	"tokyo":    "Rainy, 20°C",
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "OK")
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	var location string
	switch r.Method {
	case "GET":
		location = r.URL.Query().Get("location")
	case "POST":
		var reqBody map[string]string
		// log.Printf("DEBUG: Headers: %v", r.Header)
		// ^ Add this if debugging headers
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			log.Printf("DEBUG: JSON Decode error: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("DEBUG: Decoded body: %v", reqBody)
	// log.Printf("DEBUG: Decoded body: %v", reqBody) --> Moved to above
		location = reqBody["location"]
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if location == "" {
		http.Error(w, "Missing location parameter", http.StatusBadRequest)
		return
	}

	log.Printf("DEBUG: Looking up location: '%s'", location)
	weather, ok := weatherData[location]
	log.Printf("DEBUG: Lookup result for '%s': %v", location, ok)
	if !ok {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	response := map[string]string{
		"location": location,
		"weather":  weather,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("DEBUG: Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	} else {
		log.Printf("DEBUG: Successfully responded 200 for '%s'", location)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() { _ = conn.Close() }()

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
	var addr string
	flag := flag.NewFlagSet("weather-server", flag.ExitOnError)
	flag.StringVar(&addr, "addr", "", "address to listen on")
	_ = flag.Parse(os.Args[1:])

	if addr == "" {
		port := os.Getenv("HTTP_PORT")
		if port == "" {
			port = "8091"
		}
		addr = ":" + port
	}

	// Handle "localhost:port" or ":port"
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if host == "localhost" {
			addr = "127.0.0.1:" + port
		}
	} else if len(addr) > 0 && addr[0] != ':' && !strings.Contains(addr, ":") {
		// Assume it's just a port
		addr = ":" + addr
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/weather", weatherHandler)
	mux.HandleFunc("/ws", wsHandler)

	log.Printf("Server starting on %s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
