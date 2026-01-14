// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a mock HTTP authenticated echo server for testing.
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	port := pflag.Int("port", 8080, "Port to listen on")
	pflag.Parse()

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		apiKey := r.Header.Get("X-Api-Key")
		if apiKey != "test-api-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		defer func() { _ = r.Body.Close() }()

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(body)
	})

	addr := fmt.Sprintf(":%d", *port)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v\n", err)
		os.Exit(1)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Starting mock HTTP authed echo server on :%d\n", actualPort)
	if *port == 0 {
		fmt.Printf("%d\n", actualPort)
	}

	server := &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
