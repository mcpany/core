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

const (
	webhookBufferSize = 1000
	webhookWorkers    = 2
	webhookBatchSize  = 10
	webhookBatchWait  = 1 * time.Second
)

// WebhookAuditStore sends audit logs to a configured webhook URL.
type WebhookAuditStore struct {
	webhookURL string
	headers    map[string]string
	client     *http.Client
	queue      chan Entry
	wg         sync.WaitGroup
	done       chan struct{}
}

// NewWebhookAuditStore creates a new WebhookAuditStore.
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
			// Drain queue
			for entry := range s.queue {
				batch = append(batch, entry)
				if len(batch) >= webhookBatchSize {
					s.sendBatch(batch)
					batch = nil
				}
			}
			s.sendBatch(batch)
			return
		}
	}
}

// Write writes an audit entry to the webhook (buffered).
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
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send webhook batch: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "Webhook batch returned status: %s\n", resp.Status)
	}
}

// Read implements the Store interface.
func (s *WebhookAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for webhook audit store")
}

// Close stops the workers and drains the queue.
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
