// Package main implements a simple webhook server for testing.

package main

import (
	"bytes"
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

// main is the entry point for the Webhook Server.
// It initializes the webhook registry, sets up handlers, and starts the HTTP server.
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
					http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
					return
				}
			}
			h(w, r)
		}
	}

	http.HandleFunc("/", withVerify(handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Webhook Server on 127.0.0.1:%s", port)
	server := &http.Server{
		Addr:              "127.0.0.1:" + port,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
