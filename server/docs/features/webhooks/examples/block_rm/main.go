// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// WebhookRequest matches the data payload sent by mcpany
type WebhookRequest struct {
	Kind     int            `json:"kind"` // 1=PreCall, 2=PostCall
	ToolName string         `json:"tool_name"`
	Inputs   map[string]any `json:"inputs"`
}

// WebhookResponse matches the expected response data
type WebhookResponse struct {
	Allowed bool    `json:"allowed"`
	Status  *Status `json:"status,omitempty"`
}

// Status represents the status of the webhook response.
// It contains a code and a message.
type Status struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/validate", validateHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port
	log.Printf("Starting webhook server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received request for tool: %s, inputs: %v", req.ToolName, req.Inputs)

	// Validation Logic
	allowed := true
	message := "Allowed"

	// Iterate over inputs to check for "rm"
	// In a real scenario, you might check specific fields like "command"
	for k, v := range req.Inputs {
		if strVal, ok := v.(string); ok {
			// Check if command starts with "rm " or is exactly "rm"
			// Also checking for " rm " in case of chained commands
			cleaned := strings.TrimSpace(strVal)
			if strings.HasPrefix(cleaned, "rm ") ||
			   cleaned == "rm" ||
			   strings.Contains(cleaned, " rm ") {
				allowed = false
				message = fmt.Sprintf("Command contains restricted keyword 'rm' in input '%s'", k)
				break
			}
		}
	}

	respData := WebhookResponse{
		Allowed: allowed,
		Status: &Status{
			Code:    200,
			Message: message,
		},
	}

	respBytes, _ := json.Marshal(respData)

	// Set CloudEvents Response Headers
	w.Header().Set("Ce-Id", uuid.New().String())
	w.Header().Set("Ce-Type", "com.mcpany.webhook.response")
	w.Header().Set("Ce-Source", "/webhook/validate")
	w.Header().Set("Ce-Specversion", "1.0")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
