// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

const (
	splunkBufferSize = 1000
	splunkWorkers    = 2
)

// SplunkAuditStore sends audit logs to Splunk HTTP Event Collector.
type SplunkAuditStore struct {
	config *configv1.SplunkConfig
	client *http.Client
	queue  chan AuditEntry
	wg     sync.WaitGroup
	done   chan struct{}
}

// NewSplunkAuditStore creates a new SplunkAuditStore.
//
// config holds the configuration settings.
//
// Returns the result.
func NewSplunkAuditStore(config *configv1.SplunkConfig) *SplunkAuditStore {
	if config == nil {
		config = &configv1.SplunkConfig{}
	}
	store := &SplunkAuditStore{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		queue: make(chan AuditEntry, splunkBufferSize),
		done:  make(chan struct{}),
	}

	for i := 0; i < splunkWorkers; i++ {
		store.wg.Add(1)
		go store.worker()
	}

	return store
}

func (e *SplunkAuditStore) worker() {
	defer e.wg.Done()
	for {
		select {
		case entry, ok := <-e.queue:
			if !ok {
				return
			}
			e.send(entry)
		case <-e.done:
			// Drain queue
			for {
				select {
				case entry, ok := <-e.queue:
					if !ok {
						return
					}
					e.send(entry)
				default:
					return
				}
			}
		}
	}
}

// Write implements the AuditStore interface.
//
// _ is an unused parameter.
// entry is the entry.
//
// Returns an error if the operation fails.
func (e *SplunkAuditStore) Write(_ context.Context, entry AuditEntry) error {
	select {
	case e.queue <- entry:
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Splunk audit queue full, dropping log: %s\n", entry.ToolName)
		return fmt.Errorf("audit queue full")
	}
}

func (e *SplunkAuditStore) send(entry AuditEntry) {
	// Use background context for sending
	ctx := context.Background()

	event := map[string]interface{}{
		"time":       entry.Timestamp.Unix(),
		"host":       "mcpany",
		"source":     e.config.GetSource(),
		"sourcetype": e.config.GetSourcetype(),
		"index":      e.config.GetIndex(),
		"event":      entry,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal splunk event: %v\n", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.config.GetHecUrl(), bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create splunk request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Splunk "+e.config.GetToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send event to splunk: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Splunk HEC returned status: %d\n", resp.StatusCode)
	}
}

// Close closes the queue and waits for workers to finish.
//
// Returns an error if the operation fails.
func (e *SplunkAuditStore) Close() error {
	close(e.done)
	close(e.queue)
	e.wg.Wait()
	return nil
}
