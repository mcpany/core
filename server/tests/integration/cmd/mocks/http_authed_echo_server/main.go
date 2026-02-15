// Package main implements a mock HTTP authenticated echo server for testing.

package main

import (
	"fmt"
	"io"
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
	fmt.Printf("Starting mock HTTP authed echo server on %s\n", addr)
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
