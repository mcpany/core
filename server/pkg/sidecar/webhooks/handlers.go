// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package webhooks defines the system webhook handlers.
package webhooks
// Handle processes the markdown conversion request.
// It expects a CloudEvent with "inputs" or "result" fields containing HTML strings or structures.
//
// Summary: Handles the markdown conversion request.
//
// Parameters:
//   - w: http.ResponseWriter. The HTTP response writer.
//   - r: *http.Request. The HTTP request.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//
//	None.
//
// Side Effects:
//   - Writes the converted Markdown to the response.
// Errors:
//   - triggers relevant error states on failure.
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
// Handle processes the text truncation request.
//
// Summary: Handles the text truncation request.
//
// Parameters:
//   - w: http.ResponseWriter. The HTTP response writer.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//
//	None.
//
// Side Effects:
// Errors:
//   - None.
//   - Writes the truncated text to the response.
// Errors:
//   - triggers relevant error states on failure.
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

// Handle processes the pagination request.
//
// Summary: Handles the pagination request.
//
// Parameters:
//   - w: http.ResponseWriter. The HTTP response writer.
//   - r: *http.Request. The HTTP request.
//
// Returns:
//
//	None.
//
// Side Effects:
// Errors:
//   - None.
//   - Writes the paginated content to the response.
// Errors:
//   - triggers relevant error states on failure.
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

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			if val <= 0 {
				val = 1
			}
			page = val
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
		newInputs := paginateRecursive(val, page, pageSize)
		respData["replacement_object"] = newInputs
	} else if val, ok := data["result"]; ok {
		newResult := paginateRecursive(val, page, pageSize)
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

// ⚡ BOLT: Optimized recursive handlers to use in-place modification and avoid heavy allocations.
// Randomized Selection from Top 5 High-Impact Targets.
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
		for k, val := range v {
			v[k] = convertToMarkdown(converter, val)
		}
		return v
	case []any:
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
		// Convert to runes to handle multi-byte characters correctly
		runes := []rune(v)
		if len(runes) > maxChars {
			return string(runes[:maxChars]) + "..."
		}
		return v
	case map[string]any:
		for k, val := range v {
			v[k] = truncateRecursive(val, maxChars)
		}
		return v
	case []any:
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

		start := (page - 1) * pageSize
		end := start + pageSize

		// ⚡ BOLT: Optimized single-pass pagination.
		// Randomized Selection from Top 5 High-Impact Targets.

		startByte := len(v)
		endByte := len(v)
		totalRunes := 0

		for i := range v {
			if totalRunes == start {
				startByte = i
			}
			if totalRunes == end {
				endByte = i
			}
			totalRunes++
		}

		if start >= totalRunes {
			return fmt.Sprintf("Page %d (empty). Total length: %d", page, totalRunes)
		}

		// Calculate total pages
		totalPages := (totalRunes + pageSize - 1) / pageSize

		chunk := v[startByte:endByte]
		return fmt.Sprintf("Page %d/%d:\n%s\n(Total: %d chars)", page, totalPages, chunk, totalRunes)

	case map[string]any:
		for k, val := range v {
			v[k] = paginateRecursive(val, page, pageSize)
		}
		return v

	case []any:
		for i, val := range v {
			v[i] = paginateRecursive(val, page, pageSize)
		}
		return v
	}
	return data
}
//   - None.
// Errors:
// Side Effects:
//
//	None.
//
// Returns:
//
//   - r: *http.Request. The HTTP request.
//   - w: http.ResponseWriter. The HTTP response writer.
// Parameters:
//
// Summary: Handles the markdown conversion request.
//
// It expects a CloudEvent with "inputs" or "result" fields containing HTML strings or structures.
// Handle processes the markdown conversion request.
