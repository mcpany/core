// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main provides a mock HTTP caching server for integration testing.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	port    = flag.Int("port", 0, "The port to listen on for HTTP requests.")
	counter int64
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&counter, 1)
		_, _ = fmt.Fprintf(w, "This is a cacheable response. Call count: %d", atomic.LoadInt64(&counter))
	})

	http.HandleFunc("/reset", func(w http.ResponseWriter, _ *http.Request) {
		atomic.StoreInt64(&counter, 0)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int64{"counter": atomic.LoadInt64(&counter)})
	})

	log.Printf("Starting caching test server on port %d...", *port)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
