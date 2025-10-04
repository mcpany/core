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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/mcpxy/core/pkg/consts"
)

const openAPISpec = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Calculator API",
    "version": "1.0.0"
  },
  "paths": {
    "/add": {
      "post": {
        "summary": "Add two numbers",
        "operationId": "add",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "a": {
                    "type": "integer"
                  },
                  "b": {
                    "type": "integer"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successful operation",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "result": {
                      "type": "integer"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
`

func main() {
	port := flag.Int("port", 0, "Port to listen on. If 0, a random available port will be chosen and printed to stdout.")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on a port: %v", err)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Listening on port %d", actualPort)

	// If port was 0, print the actual chosen port to stdout so the test runner can pick it up.
	if *port == 0 {
		fmt.Printf("%d\n", actualPort) // Output port for test runner
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", consts.ContentTypeApplicationJSON)
		_, _ = w.Write([]byte(openAPISpec))
	})

	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]int
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		a, aOk := body["a"]
		b, bOk := body["b"]

		if !aOk || !bOk {
			http.Error(w, "Invalid input: 'a' and 'b' are required", http.StatusBadRequest)
			return
		}

		result := a + b
		w.Header().Set("Content-Type", consts.ContentTypeApplicationJSON)
		_ = json.NewEncoder(w).Encode(map[string]int{"result": result})
	})

	server := &http.Server{
		Handler: mux,
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
