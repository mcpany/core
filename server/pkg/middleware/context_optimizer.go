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

		// Only check successful JSON responses
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
	body *bytes.Buffer
}

func (w *responseBuffer) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *responseBuffer) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}
