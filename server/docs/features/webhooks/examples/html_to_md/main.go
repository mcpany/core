// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/google/uuid"
)

// WebhookRequest matches the data payload sent by mcpany.
type WebhookRequest struct {
	Kind     int    `json:"kind"` // 1=PreCall, 2=PostCall
	ToolName string `json:"tool_name"`
	Result   any    `json:"result"`
}

// WebhookResponse matches the expected response data.
type WebhookResponse struct {
	ReplacementObject any `json:"replacement_object,omitempty"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	http.HandleFunc("/convert", convertHandler)
	log.Println("Starting html-to-md webhook on :8082")
	server := &http.Server{
		Addr:              ":8082",
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.ListenAndServe()
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received post-call request for tool: %s", req.ToolName)

	converter := md.NewConverter("", true, nil)

	var markdown string
	var processingErr error

	switch v := req.Result.(type) {
	case string:
		markdown, processingErr = converter.ConvertString(v)
	case map[string]interface{}:
		if val, ok := v["raw"]; ok {
			if s, ok := val.(string); ok {
				markdown, processingErr = converter.ConvertString(s)
			}
		}
	}

	if processingErr != nil {
		log.Printf("Conversion failed: %v", processingErr)
		w.WriteHeader(StatusOK)
		_, _ = w.Write([]byte("{}"))
		return
	}

	if markdown == "" {
		w.WriteHeader(StatusOK)
		_, _ = w.Write([]byte("{}"))
		return
	}

	respData := WebhookResponse{
		ReplacementObject: map[string]string{
			"content": markdown,
			"format":  "markdown",
		},
	}

	respBytes, _ := json.Marshal(respData)

	w.Header().Set("Ce-Id", uuid.New().String())
	w.Header().Set("Ce-Type", "com.mcpany.webhook.response")
	w.Header().Set("Ce-Source", "/webhook/convert")
	w.Header().Set("Ce-Specversion", "1.0")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBytes)
}

// StatusOK represents the HTTP 200 OK status code.
const StatusOK = 200
