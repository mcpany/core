package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

type WebhookRequest struct {
	Uid      string          `json:"uid"`
	Kind     string          `json:"kind"`
	ToolName string          `json:"tool_name"`
	Object   json.RawMessage `json:"object"`
}

type WebhookResponse struct {
	Uid               string          `json:"uid"`
	Allowed           bool            `json:"allowed"`
	Status            *WebhookStatus  `json:"status,omitempty"`
	ReplacementObject json.RawMessage `json:"replacement_object,omitempty"`
}

type WebhookStatus struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type WebhookReview struct {
	Request  WebhookRequest   `json:"request"`
	Response *WebhookResponse `json:"response,omitempty"`
}

func main() {
	http.HandleFunc("/markdown", handleMarkdown)
	http.HandleFunc("/truncate", handleTruncate)
	http.HandleFunc("/paginate", handlePaginate)

	log.Println("Starting Webhook Server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleMarkdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var review WebhookReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req := review.Request
	resp := &WebhookResponse{
		Uid:     req.Uid,
		Allowed: true,
	}

	if req.Kind == "PostCall" && req.Object != nil {
		var obj any
		if err := json.Unmarshal(req.Object, &obj); err == nil {
			converter := md.NewConverter("", true, nil)
			newObj := convertToMarkdown(converter, obj)

			// If convertToMarkdown returned same object structure (map with "value" key if it was wrapped)
			// we need to serialize it back.
			if b, err := json.Marshal(newObj); err == nil {
				resp.ReplacementObject = b
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WebhookReview{
		Request:  req,
		Response: resp,
	})
}

func convertToMarkdown(converter *md.Converter, data any) any {
	switch v := data.(type) {
	case string:
		res, err := converter.ConvertString(v)
		if err != nil {
			return v
		}
		return res
	case map[string]any:
		newMap := make(map[string]any, len(v))
		for k, val := range v {
			newMap[k] = convertToMarkdown(converter, val)
		}
		return newMap
	case []any:
		newSlice := make([]any, len(v))
		for i, val := range v {
			newSlice[i] = convertToMarkdown(converter, val)
		}
		return newSlice
	}
	return data
}

func handleTruncate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	maxChars := 100 // default
	if m := r.URL.Query().Get("max_chars"); m != "" {
		if val, err := strconv.Atoi(m); err == nil {
			maxChars = val
		}
	}

	var review WebhookReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req := review.Request
	resp := &WebhookResponse{
		Uid:     req.Uid,
		Allowed: true,
	}

	if req.Kind == "PostCall" && req.Object != nil {
		var obj any
		if err := json.Unmarshal(req.Object, &obj); err == nil {
			newObj := truncateRecursive(obj, maxChars)
			if b, err := json.Marshal(newObj); err == nil {
				resp.ReplacementObject = b
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WebhookReview{
		Request:  req,
		Response: resp,
	})
}

func truncateRecursive(data any, maxChars int) any {
	switch v := data.(type) {
	case string:
		if len(v) > maxChars {
			return v[:maxChars] + "..."
		}
		return v
	case map[string]any:
		newMap := make(map[string]any, len(v))
		for k, val := range v {
			newMap[k] = truncateRecursive(val, maxChars)
		}
		return newMap
	case []any:
		newSlice := make([]any, len(v))
		for i, val := range v {
			newSlice[i] = truncateRecursive(val, maxChars)
		}
		return newSlice
	}
	return data
}

func handlePaginate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageSize := 1000
	if p := r.URL.Query().Get("page_size"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			pageSize = val
		}
	}

	var review WebhookReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req := review.Request
	resp := &WebhookResponse{
		Uid:     req.Uid,
		Allowed: true,
	}

	if req.Kind == "PostCall" && req.Object != nil {
		// Paginate logic needs page info.
		// Since we stripped custom inputs logic from main execution, we rely on URL params or
		// maybe the object itself acts as context?
		// Or we assume page=1 if not specified?
		// Ideally the webhook URL itself has the page param if we are browsing?
		// But the ToolExecutionRequest is stateless usually.
		// If the user tool supports pagination, it would be in inputs.
		// If independent pagination hook, maybe it modifies the OUTPUT to show only page 1,
		// and adds a "next_page_url" or similar?
		// Let's implement simple page 1 truncation for now.

		page := 1
		// We could check req.Object["page"] if it was passed through?
		// But req.Object is the RESULT of the tool.

		var obj any
		if err := json.Unmarshal(req.Object, &obj); err == nil {
			newObj := paginateRecursive(obj, page, pageSize)
			if b, err := json.Marshal(newObj); err == nil {
				resp.ReplacementObject = b
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WebhookReview{
		Request:  req,
		Response: resp,
	})
}

func paginateRecursive(data any, page, pageSize int) any {
	switch v := data.(type) {
	case string:
		runes := []rune(v)
		start := (page - 1) * pageSize
		if start >= len(runes) {
			return fmt.Sprintf("Page %d (empty). Total length: %d", page, len(runes))
		}
		end := start + pageSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[start:end])
		return fmt.Sprintf("Page %d/%d:\n%s\n(Total: %d chars)", page, (len(runes)+pageSize-1)/pageSize, chunk, len(runes))
	case map[string]any:
		// Only paginate "content" or specific fields? recursive for now
		newMap := make(map[string]any, len(v))
		for k, val := range v {
			newMap[k] = paginateRecursive(val, page, pageSize)
		}
		return newMap
	}
	return data
}
