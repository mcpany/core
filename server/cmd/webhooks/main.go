// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements the Standard Webhook Sidecar for MCP Any.
// It provides built-in handlers for common tasks like Markdown conversion,
// text truncation, and pagination.
package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcpany/core/server/pkg/sidecar/webhooks"
	webhook "github.com/standard-webhooks/standard-webhooks/libraries/go"
)

const (
	// HeaderWebhookID is the header name for the webhook ID.
	HeaderWebhookID = webhook.HeaderWebhookID
	// HeaderWebhookTimestamp is the header name for the webhook timestamp.
	HeaderWebhookTimestamp = webhook.HeaderWebhookTimestamp
	// HeaderWebhookSignature is the header name for the webhook signature.
	HeaderWebhookSignature = webhook.HeaderWebhookSignature
)

func main() {
	var (
		port   string
		secret string
	)

	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.StringVar(&secret, "secret", "", "Webhook secret for verification (or set WEBHOOK_SECRET env var)")
	flag.Parse()

	if secret == "" {
		secret = os.Getenv("WEBHOOK_SECRET")
	}

	var hook *webhook.Webhook
	if secret != "" {
		var err error
		hook, err = webhook.NewWebhook(secret)
		if err != nil {
			log.Fatalf("Failed to create webhook verifier: %v", err)
		}
		log.Printf("Webhook signature verification enabled")
	} else {
		log.Printf("Warning: Webhook signature verification disabled (no secret provided)")
	}

	reg := webhooks.NewRegistry()

	// Register Standard System Webhooks
	reg.Register("/markdown", &webhooks.MarkdownHandler{})
	reg.Register("/truncate", &webhooks.TruncateHandler{})
	reg.Register("/paginate", &webhooks.PaginateHandler{})

	log.Printf("Registered handlers: /markdown, /truncate, /paginate")

	// Main handler with routing logic
	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if h, ok := reg.Get(path); ok {
			h.Handle(w, r)
			return
		}
		http.NotFound(w, r)
	}

	// Wrapper for signature verification
	withVerify := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if hook != nil {
				// Limit body size to 1MB to prevent DoS
				body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
				if err != nil {
					http.Error(w, "Failed to read body", http.StatusInternalServerError)
					return
				}
				// Restore body for handler
				r.Body = io.NopCloser(bytes.NewReader(body))

				if err := hook.Verify(body, r.Header); err != nil {
					log.Printf("Signature verification failed: %v", err)
					http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
					return
				}
			}
			h(w, r)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", withVerify(handler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Starting Standard Webhook Sidecar on :%s", port)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case <-shutdown:
		log.Println("Starting shutdown...")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in %v: %v", 5*time.Second, err)
			if err := server.Close(); err != nil {
				log.Printf("Could not stop http server: %v", err)
			}
		}
	}
	log.Println("Server stopped")
}
