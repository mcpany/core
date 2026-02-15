// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGzipCompressionMiddleware_Flush(t *testing.T) {
	// Create a handler that flushes data in chunks
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Write part 1
		fmt.Fprintf(w, "data: part1\n\n")
		flusher.Flush()

		// Simulate delay
		time.Sleep(100 * time.Millisecond)

		// Write part 2
		fmt.Fprintf(w, "data: part2\n\n")
		flusher.Flush()
	})

	// Wrap with Gzip middleware
	gzipHandler := GzipCompressionMiddleware(handler)

	// Start a test server
	server := httptest.NewServer(gzipHandler)
	defer server.Close()

	// Create a request with Accept-Encoding: gzip
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")

	// Use a custom client with no timeout for reading stream
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %s", resp.Header.Get("Content-Encoding"))
	}

	// Read the response body as a stream
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	// We expect "data: part1" immediately.
	// Since we can't easily measure "immediately" in unit test without flaky timing,
	// we just ensure we receive both parts correctly.
	// The fact that the client returns means headers were sent (due to Flush).
	// If Flush was broken, headers might be buffered until close (for small payloads).

	// Wait, if Flush is broken, the Gzip middleware would buffer "part1" (small).
	// Then it would buffer "part2" (small).
	// Then handler returns.
	// Then middleware Closes and flushes everything.
	// So client receives everything at once.

	// To detect buffering, we can check if the response headers are received *before* the handler finishes.
	// But `client.Do` blocks until headers are received.
	// If `Flush` works, `client.Do` returns immediately after first flush.
	// If `Flush` is broken (buffering), `client.Do` might block until handler returns (200ms delay).

	// Let's use a channel to detect when `client.Do` returns.
	done := make(chan struct{})
	go func() {
		// Read all lines
		for scanner.Scan() {
			t.Logf("Received: %s", scanner.Text())
		}
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for response body")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}
}

func TestGzipCompressionMiddleware_Flush_Ordering(t *testing.T) {
	// Verify that data is flushed immediately and not buffered until the end.

	chunk1 := "chunk1"
	chunk2 := "chunk2"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("ResponseWriter does not implement http.Flusher")
			return
		}

		w.Write([]byte(chunk1))
		flusher.Flush() // Should force gzip output

		// We use a channel to coordinate with the client to prove ordering
		// But in a single process test, we can just sleep.
		time.Sleep(500 * time.Millisecond)

		w.Write([]byte(chunk2))
		// Handler returns, closing response
	})

	server := httptest.NewServer(GzipCompressionMiddleware(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("gzip reader failed: %v", err)
	}

	// Read first chunk
	buf := make([]byte, len(chunk1))

	// We expect this read to return quickly (before the 500ms sleep finishes)
	// IF flush works.
	// If flush is broken, this read will block for 500ms until handler closes.

	start := time.Now()
	n, err := io.ReadFull(gzReader, buf)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("ReadFull failed: %v", err)
	}
	if n != len(chunk1) || string(buf) != chunk1 {
		t.Errorf("Expected %q, got %q", chunk1, string(buf))
	}

	t.Logf("Read duration: %v", duration)

	if duration > 400 * time.Millisecond {
		t.Errorf("Read took too long (%v), implying flush didn't work and we waited for handler to finish", duration)
	}
}
