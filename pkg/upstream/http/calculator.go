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

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// Calculator is a mock server that simulates a calculator service.
type Calculator struct {
	server *httptest.Server
}

// NewCalculator creates a new instance of the Calculator mock server.
func NewCalculator() *Calculator {
	c := &Calculator{}
	c.server = httptest.NewServer(c.handler())
	return c
}

// handler returns the HTTP handler for the calculator service.
func (c *Calculator) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/add", c.add)
	mux.HandleFunc("/subtract", c.subtract)
	mux.HandleFunc("/multiply", c.multiply)
	mux.HandleFunc("/divide", c.divide)
	mux.HandleFunc("/power", c.power)
	mux.HandleFunc("/apikey", c.apiKey)
	mux.HandleFunc("/bearertoken", c.bearerToken)
	mux.HandleFunc("/path/add/", c.pathAdd)
	mux.HandleFunc("/form/add", c.formAdd)
	mux.HandleFunc("/template/hello", c.templateHello)
	mux.HandleFunc("/template/add", c.templateAdd)
	mux.HandleFunc("/echo", c.echo)
	return mux
}

// add handles POST /add with JSON body
func (c *Calculator) add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := data.A + data.B
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// subtract handles GET /subtract with query parameters
func (c *Calculator) subtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	a, errA := strconv.Atoi(aStr)
	b, errB := strconv.Atoi(bStr)

	if errA != nil || errB != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	result := a - b
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// multiply handles PUT /multiply with JSON body
func (c *Calculator) multiply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := data.A * data.B
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// divide handles DELETE /divide with JSON body
func (c *Calculator) divide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.B == 0 {
		http.Error(w, "division by zero", http.StatusBadRequest)
		return
	}

	result := data.A / data.B
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// power handles PATCH /power with JSON body
func (c *Calculator) power(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := 1
	for i := 0; i < data.B; i++ {
		result *= data.A
	}

	fmt.Fprintf(w, `{"result": %d}`, result)
}

// apiKey handles GET /apikey with an API key header
func (c *Calculator) apiKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("X-API-Key") != "secret" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, `{"result": "ok"}`)
}

// bearerToken handles GET /bearertoken with a Bearer token
func (c *Calculator) bearerToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token != "secret" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, `{"result": "ok"}`)
}

// pathAdd handles GET /path/add/{a}/{b} with URL path parameters
func (c *Calculator) pathAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 { // Expecting "path", "add", "a", "b"
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	a, errA := strconv.Atoi(parts[2])
	b, errB := strconv.Atoi(parts[3])

	if errA != nil || errB != nil {
		http.Error(w, "invalid path parameters", http.StatusBadRequest)
		return
	}

	result := a + b
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// formAdd handles POST /form/add with form data
func (c *Calculator) formAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	aStr := r.FormValue("a")
	bStr := r.FormValue("b")

	a, errA := strconv.Atoi(aStr)
	b, errB := strconv.Atoi(bStr)

	if errA != nil || errB != nil {
		http.Error(w, "invalid form parameters", http.StatusBadRequest)
		return
	}

	result := a + b
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// templateHello handles GET /template/hello?name=... and returns a rendered string
func (c *Calculator) templateHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	// Simple template rendering
	response := fmt.Sprintf("Hello, %s!", name)
	fmt.Fprintf(w, `{"result": "%s"}`, response)
}

// templateAdd handles POST /template/add with a templated body
func (c *Calculator) templateAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusInternalServerError)
		return
	}

	var a, b int
	_, err = fmt.Sscanf(string(body), "add %d and %d", &a, &b)
	if err != nil {
		http.Error(w, "invalid template body", http.StatusBadRequest)
		return
	}

	result := a + b
	fmt.Fprintf(w, `{"result": %d}`, result)
}

// echo handles POST /echo by returning the request body
func (c *Calculator) echo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusInternalServerError)
		return
	}
	w.Write(body)
}


// URL returns the URL of the mock server.
func (c *Calculator) URL() string {
	return c.server.URL
}

// Close closes the mock server.
func (c *Calculator) Close() {
	c.server.Close()
}
