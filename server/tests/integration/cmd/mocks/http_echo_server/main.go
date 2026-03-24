// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":0"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer func() { _ = ln.Close() }()

	fmt.Printf("LISTENING ON %s\n", ln.Addr().String())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		_, _ = w.Write(body)
	})

	server := &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
	}

	return server.Serve(ln)
}
