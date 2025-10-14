/*
 * Copyright 2025 Author(s) of MCP-XY
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

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("websocket_echo_server: Received request", "path", r.URL.Path)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket_echo_server: Failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("websocket_echo_server: Client connected")

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			slog.Error("websocket_echo_server: Error reading message", "error", err)
			break
		}
		slog.Info("websocket_echo_server: Received message", "type", messageType, "payload", string(p))
		if err := conn.WriteMessage(messageType, p); err != nil {
			slog.Error("websocket_echo_server: Error writing message", "error", err)
			break
		}
		slog.Info("websocket_echo_server: Echoed message", "type", messageType, "payload", string(p))
	}
}

// main starts the mock websocket echo server.
func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("websocket_echo_server: Failed to listen on a port", "error", err)
		os.Exit(1)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	slog.Info("websocket_echo_server: Listening on port", "port", actualPort)

	// If port was 0, print the actual chosen port to stdout so the test runner can pick it up.
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	server := &http.Server{
		Handler: mux,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to start the server
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("websocket_echo_server: Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received
	<-stop

	slog.Info("websocket_echo_server: Shutting down the server...")

	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("websocket_echo_server: Server Shutdown Failed", "error", err)
	}

	slog.Info("websocket_echo_server: Server gracefully stopped")
}
