// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package webhooks defines the system webhook handlers.
package webhooks

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// KindPostCall identifies a post-call webhook.
const KindPostCall = "PostCall"

// MarkdownHandler is a webhook handler that converts HTML content to Markdown.
// It processes incoming CloudEvents containing HTML and returns the converted Markdown.
type MarkdownHandler struct{}

// Handle processes the markdown conversion request.
// It expects a CloudEvent with "inputs" or "result" fields containing HTML strings or structures.
//
// Parameters:
//   w: The HTTP response writer.
//   r: The HTTP request.
func (h *MarkdownHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		http.Error(w, "Failed to parse CloudEvent: "+err.Error(), http.StatusBadRequest)
		return
	}

	// We expect data to be map[string]any
	var data map[string]any
	if err := event.DataAs(&data); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	respEvent := cloudevents.NewEvent()
	respEvent.SetID(uuid.New().String())
	respEvent.SetSource("https://github.com/mcpany/webhooks/markdown")
	respEvent.SetType("com.mcpany.webhook.response")
	respEvent.SetTime(time.Now())

	respData := map[string]any{
		"allowed": true,
	}

	// Check inputs or result
	if val, ok := data["inputs"]; ok {
		// Pre-Call
		converter := md.NewConverter("", true, nil)
		newInputs := convertToMarkdown(converter, val)
		respData["replacement_object"] = newInputs
	} else if val, ok := data["result"]; ok {
		// Post-Call
		converter := md.NewConverter("", true, nil)
		newResult := convertToMarkdown(converter, val)
		respData["replacement_object"] = newResult
	}

	if err := respEvent.SetData(cloudevents.ApplicationJSON, respData); err != nil {
		http.Error(w, "Failed to set response data", http.StatusInternalServerError)
		return
	}

	// Write response event
	w.Header().Set("Content-Type", "application/cloudevents+json")
	_ = json.NewEncoder(w).Encode(respEvent)
}

// TruncateHandler is a webhook handler that truncates long strings to a specified length.
// It processes incoming CloudEvents and truncates strings in "inputs" or "result" fields.
// The maximum characters can be specified via the "max_chars" query parameter (default 100).
type TruncateHandler struct{}

// Handle processes the text truncation request.
//
// Parameters:
//   w: The HTTP response writer.
//   r: The HTTP request.
func (h *TruncateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	maxChars := 100
	if m := r.URL.Query().Get("max_chars"); m != "" {
		if val, err := strconv.Atoi(m); err == nil {
			if val <= 0 {
				val = 1
			}
			if val > 100000 {
				val = 100000
			}
			maxChars = val
		}
	}

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		http.Error(w, "Failed to parse CloudEvent: "+err.Error(), http.StatusBadRequest)
		return
	}

	var data map[string]any
	if err := event.DataAs(&data); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	respEvent := cloudevents.NewEvent()
	respEvent.SetID(uuid.New().String())
	respEvent.SetSource("https://github.com/mcpany/webhooks/truncate")
	respEvent.SetType("com.mcpany.webhook.response")
	respEvent.SetTime(time.Now())

	respData := map[string]any{
		"allowed": true,
	}

	if val, ok := data["inputs"]; ok {
		newInputs := truncateRecursive(val, maxChars)
		respData["replacement_object"] = newInputs
	} else if val, ok := data["result"]; ok {
		newResult := truncateRecursive(val, maxChars)
		respData["replacement_object"] = newResult
	}

	if err := respEvent.SetData(cloudevents.ApplicationJSON, respData); err != nil {
		http.Error(w, "Failed to set response data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/cloudevents+json")
	_ = json.NewEncoder(w).Encode(respEvent)
}

// PaginateHandler is a webhook handler that splits long strings into pages.
// It processes incoming CloudEvents and paginates strings in "inputs" or "result" fields.
// The page size can be specified via the "page_size" query parameter (default 1000).
type PaginateHandler struct{}

// Handle processes the pagination request.
//
// Parameters:
//   w: The HTTP response writer.
//   r: The HTTP request.
func (h *PaginateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageSize := 1000
	if p := r.URL.Query().Get("page_size"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			if val <= 0 {
				val = 1
			}
			if val > 10000 {
				val = 10000
			}
			pageSize = val
		}
	}

	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		http.Error(w, "Failed to parse CloudEvent: "+err.Error(), http.StatusBadRequest)
		return
	}

	var data map[string]any
	if err := event.DataAs(&data); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	respEvent := cloudevents.NewEvent()
	respEvent.SetID(uuid.New().String())
	respEvent.SetSource("https://github.com/mcpany/webhooks/paginate")
	respEvent.SetType("com.mcpany.webhook.response")
	respEvent.SetTime(time.Now())

	respData := map[string]any{
		"allowed": true,
	}

	if val, ok := data["inputs"]; ok {
		newInputs := paginateRecursive(val, 1, pageSize)
		respData["replacement_object"] = newInputs
	} else if val, ok := data["result"]; ok {
		newResult := paginateRecursive(val, 1, pageSize)
		respData["replacement_object"] = newResult
	}

	if err := respEvent.SetData(cloudevents.ApplicationJSON, respData); err != nil {
		http.Error(w, "Failed to set response data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/cloudevents+json")
	if err := json.NewEncoder(w).Encode(respEvent); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Helpers

func convertToMarkdown(converter *md.Converter, data any) any {
	switch v := data.(type) {
	case string:
		if len(v) > 1024*1024 {
			return "Error: Input too large"
		}
		res, err := converter.ConvertString(v)
		if err != nil {
			return v
		}
		return res
	case map[string]any:
		// ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets
		// Optimized to mutate map in-place to avoid O(N) memory allocation for deep copies.
		for k, val := range v {
			v[k] = convertToMarkdown(converter, val)
		}
		return v
	case []any:
		// ⚡ BOLT: Optimized to mutate slice in-place.
		for i, val := range v {
			v[i] = convertToMarkdown(converter, val)
		}
		return v
	}
	return data
}

func truncateRecursive(data any, maxChars int) any {
	switch v := data.(type) {
	case string:
		if len(v) <= maxChars {
			return v
		}
		// ⚡ BOLT: Fixed UTF-8 slicing bug and optimized memory by scanning runes instead of allocating slice.
		var count int
		for i := range v {
			if count >= maxChars {
				return v[:i] + "..."
			}
			count++
		}
		return v
	case map[string]any:
		// ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets
		// Optimized to mutate map in-place to avoid O(N) memory allocation for deep copies.
		for k, val := range v {
			v[k] = truncateRecursive(val, maxChars)
		}
		return v
	case []any:
		// ⚡ BOLT: Optimized to mutate slice in-place.
		for i, val := range v {
			v[i] = truncateRecursive(val, maxChars)
		}
		return v
	}
	return data
}

func paginateRecursive(data any, page, pageSize int) any {
	switch v := data.(type) {
	case string:
		if len(v) > 1024*1024 {
			return "Error: Input too large"
		}
		// ⚡ BOLT: Optimized to avoid large []rune allocation.
		var startByte, endByte, totalRunes int
		startTarget := (page - 1) * pageSize
		endTarget := startTarget + pageSize
		foundStart := false

		for i := range v {
			if totalRunes == startTarget {
				startByte = i
				foundStart = true
			}
			if totalRunes == endTarget {
				endByte = i
			}
			totalRunes++
		}

		if !foundStart {
			return fmt.Sprintf("Page %d (empty). Total length: %d", page, totalRunes)
		}

		if endByte == 0 {
			endByte = len(v)
		}

		chunk := v[startByte:endByte]
		return fmt.Sprintf("Page %d/%d:\n%s\n(Total: %d chars)", page, (totalRunes+pageSize-1)/pageSize, chunk, totalRunes)
	case map[string]any:
		// ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets
		// Optimized to mutate map in-place to avoid O(N) memory allocation for deep copies.
		for k, val := range v {
			v[k] = paginateRecursive(val, page, pageSize)
		}
		return v
	case []any:
		// ⚡ BOLT: Added missing slice recursion and optimized to mutate in-place.
		for i, val := range v {
			v[i] = paginateRecursive(val, page, pageSize)
		}
		return v
	}
	return data
}
