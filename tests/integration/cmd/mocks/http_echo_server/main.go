/*
 * Copyright 2025 Author(s) of MCPXY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/mcpxy/core/pkg/consts"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("http_echo_server: Received request", "method", r.Method, "path", r.URL.Path)
	if r.Method == http.MethodPost && r.URL.Path == "/echo" {
		bodyBytes, errRead := io.ReadAll(r.Body)
		if errRead != nil {
			slog.Error("http_echo_server: Error reading request body", "error", errRead)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", consts.ContentTypeApplicationJSON) // Assume JSON echo
		w.WriteHeader(http.StatusOK)
		if _, errWrite := w.Write(bodyBytes); errWrite != nil {
			slog.Error("http_echo_server: Error writing response body", "error", errWrite)
		}
		slog.Info("http_echo_server: Responded to POST /echo", "bytes", len(bodyBytes))
		return
	}
	slog.Warn("http_echo_server: Path not found or method not allowed", "path", r.URL.Path, "method", r.Method)
	http.NotFound(w, r)
}

// main starts the mock HTTP echo server.
func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("http_echo_server: Failed to listen on a port", "error", err)
		os.Exit(1)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	slog.Info("http_echo_server: Listening on port", "port", actualPort)

	// If port was 0, print the actual chosen port to stdout so the test runner can pick it up.
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	server := &http.Server{
		Addr:    addr, // This will be overridden by the listener below for port 0 case
		Handler: mux,
	}

	// Use the listener created above, which handles the random port assignment if port was 0
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		slog.Error("http_echo_server: Server failed", "error", err)
		os.Exit(1)
	}
	slog.Info("http_echo_server: Server shut down.")
}
