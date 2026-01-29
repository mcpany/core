// Package main provides a mock HTTP echo server for integration testing.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcpany/core/server/pkg/consts"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("http_echo_server: Received request", "method", r.Method, "path", r.URL.Path)
	if r.Method != http.MethodPost {
		slog.Warn("http_echo_server: Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, errRead := io.ReadAll(r.Body)
	if errRead != nil {
		slog.Error("http_echo_server: Error reading request body", "error", errRead)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", consts.ContentTypeApplicationJSON) // Assume JSON echo
	w.WriteHeader(http.StatusOK)
	if _, errWrite := w.Write(bodyBytes); errWrite != nil {
		slog.Error("http_echo_server: Error writing response body", "error", errWrite)
	}
	slog.Info("http_echo_server: Responded to POST /echo", "bytes", len(bodyBytes))
}

// main starts the mock HTTP echo server.
func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		slog.Error("http_echo_server: Failed to listen on a port", "error", err)
		os.Exit(1)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	slog.Info("http_echo_server: Listening on port", "port", actualPort)

	// If port was 0, print the actual chosen port to stdout so the test runner can pick it up.
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})

	server := &http.Server{
		Addr:              addr, // This will be overridden by the listener below for port 0 case
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to start the server
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("http_echo_server: Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received
	<-stop

	slog.Info("http_echo_server: Shutting down the server...")

	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("http_echo_server: Server Shutdown Failed", "error", err)
	}

	slog.Info("http_echo_server: Server gracefully stopped")
}
