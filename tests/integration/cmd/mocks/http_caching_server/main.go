/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

var (
	port    = flag.Int("port", 0, "The port to listen on for HTTP requests.")
	counter int64
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&counter, 1)
		fmt.Fprintf(w, "This is a cacheable response. Call count: %d", atomic.LoadInt64(&counter))
	})

	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreInt64(&counter, 0)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int64{"counter": atomic.LoadInt64(&counter)})
	})

	log.Printf("Starting caching test server on port %d...", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
