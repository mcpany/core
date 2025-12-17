// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a simple webhook server for testing.
package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mcpany/core/cmd/webhooks/hooks"
	webhook "github.com/standard-webhooks/standard-webhooks/libraries/go"
)

const (
	HeaderWebhookID        = webhook.HeaderWebhookID
	HeaderWebhookTimestamp = webhook.HeaderWebhookTimestamp
	HeaderWebhookSignature = webhook.HeaderWebhookSignature
)

func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	var hook *webhook.Webhook
	if secret != "" {
		var err error
		hook, err = webhook.NewWebhook(secret)
		if err != nil {
			log.Fatalf("Failed to create webhook verify: %v", err)
		}
	}

	reg := hooks.NewRegistry()
	// Simulate loading from config
	// In real world, we would read a directory of yaml/proto files
	// For now, we manually register the known system hooks
	reg.Register("/markdown", &hooks.MarkdownHandler{})
	reg.Register("/truncate", &hooks.TruncateHandler{})
	reg.Register("/paginate", &hooks.PaginateHandler{})

	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if h, ok := reg.Get(path); ok {
			h.Handle(w, r)
			return
		}
		http.NotFound(w, r)
	}

	withVerify := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if hook != nil {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read body", http.StatusInternalServerError)
					return
				}
				// Restore body for handler
				r.Body = io.NopCloser(bytes.NewReader(body))

				if err := hook.Verify(body, r.Header); err != nil {
					http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
					return
				}
			}
			h(w, r)
		}
	}

	http.HandleFunc("/", withVerify(handler))

	log.Println("Starting Webhook Server on :8080")
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
