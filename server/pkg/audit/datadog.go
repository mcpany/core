// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package audit provides implementations of audit stores.
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

	configv1 "github.com/mcpany/core/proto/config/v1"
)

const (
	datadogBufferSize = 1000
	datadogWorkers    = 2
	datadogBatchSize  = 10
	datadogBatchWait  = 1 * time.Second
)

// DatadogAuditStore - Auto-generated documentation.
//
// Summary: DatadogAuditStore sends audit logs to Datadog.
//
// Fields:
//   - Various fields for DatadogAuditStore.
type DatadogAuditStore struct {
	config *configv1.DatadogConfig
	client *http.Client
	url    string
	queue  chan Entry
	wg     sync.WaitGroup
	done   chan struct{}
}

// NewDatadogAuditStore creates a new DatadogAuditStore.
//
// Summary: Initializes a new DatadogAuditStore with background workers.
//
// Parameters:
//   - config: *configv1.DatadogConfig. The Datadog configuration.
//
// Returns:
//   - *DatadogAuditStore: The initialized store.
//
// Side Effects:
//   - Starts background workers to process the log queue.
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
		queue: make(chan Entry, datadogBufferSize),
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
	var batch []Entry
	ticker := time.NewTicker(datadogBatchWait)
	defer ticker.Stop()

	for {
		select {
		case entry, ok := <-e.queue:
			if !ok {
				e.sendBatch(batch)
				return
			}
			batch = append(batch, entry)
			if len(batch) >= datadogBatchSize {
				e.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				e.sendBatch(batch)
				batch = nil
			}
		case <-e.done:
			// Drain queue
			for entry := range e.queue {
				batch = append(batch, entry)
				if len(batch) >= datadogBatchSize {
					e.sendBatch(batch)
					batch = nil
				}
			}
			e.sendBatch(batch)
			return
		}
	}
}

// Write implements the Store interface.
//
// Summary: Queues an audit entry for sending to Datadog.
//
// Parameters:
//   - _: context.Context. Unused.
//   - entry: Entry. The audit entry to log.
//
// Returns:
//   - error: An error if the queue is full.
//
// Errors:
//   - Returns "audit queue full" if the buffer is exceeded.
//
// Side Effects:
//   - Sends the entry to a buffered channel.
func (e *DatadogAuditStore) Write(_ context.Context, entry Entry) error {
	select {
	case e.queue <- entry:
		return nil
	default:
		// Queue full
		fmt.Fprintf(os.Stderr, "Datadog audit queue full, dropping log: %s\n", entry.ToolName)
		return fmt.Errorf("audit queue full")
	}
}

func (e *DatadogAuditStore) sendBatch(batch []Entry) {
	if len(batch) == 0 {
		return
	}

	ddLogs := make([]map[string]interface{}, 0, len(batch))
	for _, entry := range batch {
		ddLog := map[string]interface{}{
			"ddsource": "mcpany",
			"service":  e.config.GetService(),
			"message":  entry,
			"ddtags":   e.config.GetTags(),
		}
		ddLogs = append(ddLogs, ddLog)
	}

	payload, err := json.Marshal(ddLogs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal datadog log batch: %v\n", err)
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", e.url, bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create datadog request: %v\n", err)
		return
	}

	req.Header.Set("DD-API-KEY", e.config.GetApiKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send log batch to datadog: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Datadog API returned status: %d\n", resp.StatusCode)
	}
}


// Read implements the Store interface. Summary: Reads audit entries (Not implemented). Parameters: - _: context.Context. Unused. - _: Filter. Unused. Returns: - []Entry: Nil. - error: Always returns "not implemented".
//
// Summary: Read implements the Store interface. Summary: Reads audit entries (Not implemented). Parameters: - _: context.Context. Unused. - _: Filter. Unused. Returns: - []Entry: Nil. - error: Always returns "not implemented".
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - _ (Filter): The _ parameter used in the operation.
//
// Returns:
//   - ([]Entry): The resulting []Entry object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (e *DatadogAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for datadog audit store")
}

// Close - Auto-generated documentation.
//
// Summary: Close closes the queue and waits for workers to finish.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (e *DatadogAuditStore) Close() error {
	if e.done != nil {
		close(e.done)
	}
	if e.queue != nil {
		close(e.queue)
	}
	e.wg.Wait()
	return nil
}
