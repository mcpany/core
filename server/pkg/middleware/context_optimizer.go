// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

// ContextOptimizer optimises the context size of responses.
type ContextOptimizer struct {
	MaxChars int
}

// NewContextOptimizer creates a new ContextOptimizer.
func NewContextOptimizer(maxChars int) *ContextOptimizer {
	return &ContextOptimizer{
		MaxChars: maxChars,
	}
}

// Middleware returns the middleware handler.
func (co *ContextOptimizer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &responseBuffer{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		c.Next()

		// If we didn't buffer, it means we passed through the original writer (streaming).
		// In that case, we don't need to do anything here.
		if !w.isBuffering() {
			return
		}

		// Only check successful JSON responses
		// Note: w.shouldBuffer should already cover the content-type check, but we double check for safety
		// and also check status code which might not have been available at first write (though it should be).
		contentType := w.Header().Get("Content-Type")
		if w.Status() == http.StatusOK && (contentType == "application/json" || strings.HasPrefix(contentType, "application/json;")) {
			bodyBytes := w.body.Bytes()

			// Fast path: check if we need to modify anything using gjson
			// This avoids expensive full unmarshaling if no truncation is needed
			needsModification := false

			// Check if result.content exists and is an array
			resultContent := gjson.GetBytes(bodyBytes, "result.content")
			if resultContent.IsArray() {
				resultContent.ForEach(func(_, value gjson.Result) bool {
					text := value.Get("text")
					// Use len(text.String()) to check length.
					// Note: gjson text.String() handles escape characters correctly for length check
					if text.Exists() && len(text.String()) > co.MaxChars {
						needsModification = true
						return false // stop iteration
					}
					return true // continue
				})
			}

			if needsModification {
				var resp map[string]interface{}
				var json = jsoniter.ConfigCompatibleWithStandardLibrary
				if err := json.Unmarshal(bodyBytes, &resp); err == nil {
					// Look for content.text deep in the structure
					// Support MCP "content" array in result
					modified := false
					if result, ok := resp["result"].(map[string]interface{}); ok {
						if content, ok := result["content"].([]interface{}); ok {
							for i, item := range content {
								if itemMap, ok := item.(map[string]interface{}); ok {
									if text, ok := itemMap["text"].(string); ok {
										if len(text) > co.MaxChars {
											itemMap["text"] = text[:co.MaxChars] + fmt.Sprintf("...[TRUNCATED %d chars]", len(text)-co.MaxChars)
											content[i] = itemMap
											modified = true
										}
									}
								}
							}
							if modified {
								result["content"] = content
								resp["result"] = result
								newBody, _ := json.Marshal(resp)
								// Reset buffer and write new body
								if _, err := w.ResponseWriter.Write(newBody); err != nil {
									// Ignore error
									_ = err
								}
								return
							}
						}
					}
				}
			}
		}

		// If not modified, write original body
		if _, err := w.ResponseWriter.Write(w.body.Bytes()); err != nil {
			_ = err
		}
	}
}

type responseBuffer struct {
	gin.ResponseWriter
	body         *bytes.Buffer
	shouldBuffer *bool
}

func (w *responseBuffer) isBuffering() bool {
	// If explicitly set to false, then false.
	if w.shouldBuffer != nil && !*w.shouldBuffer {
		return false
	}
	// If true or nil (default), we assume buffering for now, OR we haven't written yet.
	// But if we haven't written yet, body is empty, so Middleware will write empty body which is fine.
	// However, if we return true here, Middleware will try to write w.body.Bytes().
	// If we haven't written anything, w.body is empty.
	return true
}

func (w *responseBuffer) checkBuffer() {
	if w.shouldBuffer == nil {
		ct := w.Header().Get("Content-Type")
		// We only buffer application/json.
		should := ct == "application/json" || strings.HasPrefix(ct, "application/json;")
		w.shouldBuffer = &should
	}
}

// Write writes the data to the buffer or the underlying ResponseWriter.
func (w *responseBuffer) Write(b []byte) (int, error) {
	w.checkBuffer()

	if *w.shouldBuffer {
		return w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// WriteString writes the string data to the buffer or the underlying ResponseWriter.
func (w *responseBuffer) WriteString(s string) (int, error) {
	w.checkBuffer()

	if *w.shouldBuffer {
		return w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}
