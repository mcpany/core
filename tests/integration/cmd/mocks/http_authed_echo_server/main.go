/*
 * Copyright 2025 Author(s) of MCPX
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
	"log"
	"net/http"
)

const (
	expectedAuthHeader = "X-Api-Key"
	expectedAuthValue  = "test-api-key"
)

func main() {
	port := flag.Int("port", 0, "Port to listen on")
	flag.Parse()

	if *port == 0 {
		log.Fatal("port is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/echo", echoHandler)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("INFO http_authed_echo_server: Listening on port port=%d", *port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("INFO http_authed_echo_server: Received request method=%s path=%s", r.Method, r.URL.Path)

	if r.Header.Get(expectedAuthHeader) != expectedAuthValue {
		log.Printf("ERROR http_authed_echo_server: Unauthorized request. Missing or invalid API key.")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		log.Printf("ERROR http_authed_echo_server: could not write response: %v", err)
	}
	log.Printf("INFO http_authed_echo_server: Responded to POST /echo bytes=%d", len(body))
}
