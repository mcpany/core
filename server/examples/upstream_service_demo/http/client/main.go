// Package main implements a demo HTTP client for the upstream service.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://localhost:8080", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("Current time: %s, Timezone: %s\n", result["current_time"], result["timezone"])
	return nil
}
