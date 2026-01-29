// Package main implements a mock OpenWeatherMap server.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/consts"
)

const openAPISpec = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Weather API",
    "version": "1.0.0"
  },
  "paths": {
    "/weather": {
      "get": {
        "summary": "Get weather for a location",
        "operationId": "getWeather",
        "parameters": [
          {
            "name": "location",
            "in": "query",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful operation",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "weather": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
`

var weatherData = map[string]string{
	"new york": "Sunny, 25°C",
	"london":   "Cloudy, 15°C",
	"tokyo":    "Rainy, 20°C",
}

func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on a port: %v", err)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Listening on port %d", actualPort)

	// If port was 0, print the actual chosen port to stdout so the test runner can pick it up.
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", consts.ContentTypeApplicationJSON)
		_, _ = w.Write([]byte(openAPISpec))
	})

	mux.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		location := r.URL.Query().Get("location")
		if location == "" {
			http.Error(w, "Missing location parameter", http.StatusBadRequest)
			return
		}
		weather, ok := weatherData[location]
		if !ok {
			http.Error(w, "Location not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", consts.ContentTypeApplicationJSON)
		_ = json.NewEncoder(w).Encode(map[string]string{"weather": weather})
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
