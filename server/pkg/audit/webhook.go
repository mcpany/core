// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// NewWebhookAuditStore creates a new WebhookAuditStore.
//
// Summary: Creates a new webhook audit store.
//
// Parameters:
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - webhookURL (string): The URL to send audit logs to.
//   - headers (map[string]string): Additional headers to send with the request.
//
// Returns:
//   - *WebhookAuditStore: A new WebhookAuditStore instance.
//
// Side Effects:
//   - Starts background workers.
// Errors:
//   - triggers relevant error states on failure.
func NewWebhookAuditStore(webhookURL string, headers map[string]string) *WebhookAuditStore {
	store := &WebhookAuditStore{
		webhookURL: webhookURL,
		headers:    headers,
		client:     &http.Client{Timeout: 10 * time.Second},
		queue:      make(chan Entry, webhookBufferSize),
		done:       make(chan struct{}),
	}

	for i := 0; i < webhookWorkers; i++ {
		store.wg.Add(1)
		go store.worker()
	}

	return store
}

func (s *WebhookAuditStore) worker() {
	defer s.wg.Done()
	var batch []Entry
	ticker := time.NewTicker(webhookBatchWait)
	defer ticker.Stop()

	for {
		select {
		case entry, ok := <-s.queue:
			if !ok {
				s.sendBatch(batch)
				return
			}
			batch = append(batch, entry)
			if len(batch) >= webhookBatchSize {
				s.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.sendBatch(batch)
				batch = nil
			}
		case <-s.done:
// Write writes an audit entry to the webhook (buffered).
//
// Summary: Queues an audit entry for sending.
//
// Parameters:
//   - _ (context.Context): Unused context.
//   - entry (Entry): The audit entry to write.
//
// Returns:
//   - error: An error if the queue is full.
//
// Side Effects:
//   - Queues the entry for processing.
// Errors:
//   - triggers relevant error states on failure.
func (s *WebhookAuditStore) Write(_ context.Context, entry Entry) error {
	select {
	case s.queue <- entry:
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Webhook audit queue full, dropping log: %s\n", entry.ToolName)
		return fmt.Errorf("audit queue full")
	}
}

func (s *WebhookAuditStore) sendBatch(batch []Entry) {
	if len(batch) == 0 {
		return
	}

	payload, err := json.Marshal(batch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit batch: %v\n", err)
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", s.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create webhook request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.headers {
// Read implements the Store interface.
//
// Summary: Reads audit logs (not implemented for webhook store).
//
// Parameters:
// Close stops the workers and drains the queue.
//
// Summary: Gracefully shuts down the webhook store.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: Always nil.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Side Effects:
//   - Stops background workers and drains the queue.
// Errors:
//   - triggers relevant error states on failure.
func (s *WebhookAuditStore) Close() error {
	if s.done != nil {
		close(s.done)
	}
	if s.queue != nil {
		close(s.queue)
	}
	s.wg.Wait()
	return nil
}
