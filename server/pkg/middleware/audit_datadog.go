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
	datadogBufferSize = 1000
	datadogWorkers    = 2
)

// DatadogAuditStore sends audit logs to Datadog.
type DatadogAuditStore struct {
	config *configv1.DatadogConfig
	client *http.Client
	url    string
	queue  chan AuditEntry
	wg     sync.WaitGroup
	done   chan struct{}
}

// NewDatadogAuditStore creates a new DatadogAuditStore.
func NewDatadogAuditStore(config *configv1.DatadogConfig) *DatadogAuditStore {
	if config == nil {
		config = &configv1.DatadogConfig{}
	}
	site := config.GetSite()
	if site == "" {
		site = "datadoghq.com"
	}
	url := fmt.Sprintf("https://http-intake.logs.%s/api/v2/logs", site)

	store := &DatadogAuditStore{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url:   url,
		queue: make(chan AuditEntry, datadogBufferSize),
		done:  make(chan struct{}),
	}

	for i := 0; i < datadogWorkers; i++ {
		store.wg.Add(1)
		go store.worker()
	}

	return store
}

func (e *DatadogAuditStore) worker() {
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
func (e *DatadogAuditStore) Write(ctx context.Context, entry AuditEntry) error {
	select {
	case e.queue <- entry:
		return nil
	default:
		// Queue full
		fmt.Fprintf(os.Stderr, "Datadog audit queue full, dropping log: %s\n", entry.ToolName)
		return fmt.Errorf("audit queue full")
	}
}

func (e *DatadogAuditStore) send(entry AuditEntry) {
	ctx := context.Background()

	ddLog := map[string]interface{}{
		"ddsource": "mcpany",
		"service":  e.config.GetService(),
		"message":  entry,
		"ddtags":   e.config.GetTags(),
	}

	payload, err := json.Marshal(ddLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal datadog log: %v\n", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.url, bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create datadog request: %v\n", err)
		return
	}

	req.Header.Set("DD-API-KEY", e.config.GetApiKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send log to datadog: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Datadog API returned status: %d\n", resp.StatusCode)
	}
}

// Close closes the queue and waits for workers to finish.
func (e *DatadogAuditStore) Close() error {
	close(e.done)
	close(e.queue)
	e.wg.Wait()
	return nil
}
