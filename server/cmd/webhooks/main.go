// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a simple webhook server for testing.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

// Config holds the configuration for the webhook server.
type Config struct {
	WebhookSecret string
	InsecureMode  bool
}

// NewHandler creates a new HTTP handler for the webhook server.
func NewHandler(cfg Config) (http.Handler, error) {
	var hook *webhook.Webhook
	if cfg.WebhookSecret != "" {
		var err error
		hook, err = webhook.NewWebhook(cfg.WebhookSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to create webhook verify: %w", err)
		}
	} else if !cfg.InsecureMode {
		return nil, fmt.Errorf("WEBHOOK_SECRET is required unless insecure mode is enabled")
	}

	reg := webhooks.NewRegistry()
	// Simulate loading from config
	// In real world, we would read a directory of yaml/proto files
	// For now, we manually register the known system hooks
	reg.Register("/markdown", &webhooks.MarkdownHandler{})
	reg.Register("/truncate", &webhooks.TruncateHandler{})
	reg.Register("/paginate", &webhooks.PaginateHandler{})

	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if h, ok := reg.Get(path); ok {
			h.Handle(w, r)
			return
		}
		http.NotFound(w, r)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Limit body size to 1MB to prevent DoS
		// We do this BEFORE verification to protect the verification logic (and everything else)
		// from reading too much.
		// Use MaxBytesReader to enforce the limit and return 413 if exceeded.
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

		if hook != nil {
			// We need to read the body for verification.
			// MaxBytesReader will handle the limit.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				// If MaxBytesReader hit the limit, it returns an error.
				// However, http.MaxBytesReader documentation says it stops reading.
				// We should check if it's "request body too large".
				// But io.ReadAll returning error usually means 500 or 413 if MaxBytesReader logic kicked in.
				// The MaxBytesReader sets a flag on response writer to close connection?
				// Actually, checking err is enough.
				// "The http.MaxBytesReader ... returns an error when the limit is reached"
				// Standard-webhooks hook.Verify takes []byte.

				// If we failed to read (e.g. too large), we can't verify.
				http.Error(w, "Failed to read body or body too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Restore body for handler
			r.Body = io.NopCloser(bytes.NewReader(body))

			if err := hook.Verify(body, r.Header); err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
		}

		handler(w, r)
	})

	return mux, nil
}

// main is the entry point for the Webhook Server.
// It initializes the webhook registry, sets up handlers, and starts the HTTP server.
func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	insecure := os.Getenv("MCPANY_INSECURE_DEV_MODE") == "true"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler, err := NewHandler(Config{
		WebhookSecret: secret,
		InsecureMode:  insecure,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting Webhook Server on 127.0.0.1:%s", port)
	server := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
