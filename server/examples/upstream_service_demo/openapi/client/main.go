// Package main implements an OpenAPI client demo.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read body: %v", err)
		return
	}

	fmt.Println(string(body))
}
