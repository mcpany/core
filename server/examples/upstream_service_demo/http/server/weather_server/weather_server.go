// Package main implements the weather server for the upstream service demo.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			log.Printf("DEBUG: JSON Decode error: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("DEBUG: Decoded body: %v", reqBody)
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var reqBody map[string]string
		if err := json.Unmarshal(msg, &reqBody); err != nil {
			log.Printf("json error: %v", err)
			break
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
			break
		}
	}
}

func main() {
	if err := run(os.Args, make(chan os.Signal, 1), nil); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(args []string, stop chan os.Signal, ready chan<- string) error {
	fs := flag.NewFlagSet("weather-server", flag.ContinueOnError)
	port := fs.String("port", "8080", "Port to listen on")
	var addr string
	fs.StringVar(&addr, "addr", "", "Address to listen on (overrides port if set)")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/weather", weatherHandler)
	mux.HandleFunc("/ws", wsHandler)

	listenAddr := "127.0.0.1:" + *port
	if addr != "" {
		listenAddr = addr
	}

	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	if ready != nil {
		ready <- ln.Addr().String()
	}

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Starting weather server on %s", ln.Addr())
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down weather server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
