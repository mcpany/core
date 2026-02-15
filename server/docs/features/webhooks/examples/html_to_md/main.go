package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/google/uuid"
)

// WebhookRequest matches the data payload sent by mcpany
type WebhookRequest struct {
	Kind     int            `json:"kind"` // 1=PreCall, 2=PostCall
	ToolName string         `json:"tool_name"`
	Result   any            `json:"result"`
}

// WebhookResponse matches the expected response data
type WebhookResponse struct {
	ReplacementObject any `json:"replacement_object,omitempty"`
}

func main() {
	http.HandleFunc("/convert", convertHandler)
	log.Println("Starting html-to-md webhook on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
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
	defer r.Body.Close()

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received post-call request for tool: %s", req.ToolName)

	converter := md.NewConverter("", true, nil)

	// We expect result to be a map or string potentially?
	// The prompt says "gemini cli to grab a webpage, (upstream returns html), then webhook converts to markdown".
	// Usually the result from an HTTP tool is the body (string) or JSON.
	// If it's raw HTML string:
	var markdown string
	var processingErr error

	// Handle different result types
	switch v := req.Result.(type) {
	case string:
		// Direct HTML string
		markdown, processingErr = converter.ConvertString(v)
	case map[string]interface{}:
		// Maybe inside a "content" field? Or "raw"?
		if val, ok := v["raw"]; ok {
			if s, ok := val.(string); ok {
				markdown, processingErr = converter.ConvertString(s)
			}
		}
		// If generic JSON, we might not want to convert.
	}

	if processingErr != nil {
		log.Printf("Conversion failed: %v", processingErr)
		// Return original if failure? Or empty?
		// We'll just return original (no replacement)
		w.WriteHeader(StatusOK)
		w.Write([]byte("{}"))
		return
	}

	if markdown == "" {
		// Nothing converted or empty
		w.WriteHeader(StatusOK)
		w.Write([]byte("{}"))
		return
	}

	respData := WebhookResponse{
		ReplacementObject: map[string]string{
			"content": markdown,
			"format": "markdown",
		},
	}

	respBytes, _ := json.Marshal(respData)

	// CloudEvents Headers
	w.Header().Set("Ce-Id", uuid.New().String())
	w.Header().Set("Ce-Type", "com.mcpany.webhook.response")
	w.Header().Set("Ce-Source", "/webhook/convert")
	w.Header().Set("Ce-Specversion", "1.0")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

// StatusOK represents the HTTP 200 OK status code.
const StatusOK = 200
