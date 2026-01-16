// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ContextOptimizer optimises the context size of responses.
// It reduces token usage by truncating large text fields in the response body.
type ContextOptimizer struct {
	MaxChars int
}

// NewContextOptimizer creates a new ContextOptimizer.
// maxChars specifies the maximum number of characters to keep in text fields.
func NewContextOptimizer(maxChars int) *ContextOptimizer {
	return &ContextOptimizer{
		MaxChars: maxChars,
	}
}

var bufferPool = sync.Pool{
	New: func() any {
		return &responseBuffer{
			body: &bytes.Buffer{},
		}
	},
}

// Middleware returns a Gin middleware handler that optimizes response context.
// It intercepts JSON responses and truncates text fields that exceed the configured maximum length.
func (co *ContextOptimizer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := bufferPool.Get().(*responseBuffer)
		w.ResponseWriter = c.Writer
		w.body.Reset()
		w.checked = false
		w.shouldBuffer = false

		originalWriter := c.Writer
		c.Writer = w

		defer func() {
			c.Writer = originalWriter
			w.ResponseWriter = nil // Avoid holding reference
			// If buffer is too large (>1MB), replace it with a new one to release memory
			if w.body.Cap() > 1024*1024 {
				w.body = &bytes.Buffer{}
			}
			bufferPool.Put(w)
		}()

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

			type modification struct {
				start       int
				end         int
				replacement []byte
			}
			var mods []modification

			// Check if result.content exists and is an array
			resultContent := gjson.GetBytes(bodyBytes, "result.content")
			if resultContent.IsArray() {
				resultContent.ForEach(func(_, value gjson.Result) bool {
					text := value.Get("text")
					// Use len(text.String()) to check length.
					// Note: gjson text.String() handles escape characters correctly for length check
					if text.Exists() {
						str := text.String()
						if len(str) > co.MaxChars {
							// Found a text that needs truncation
							truncated := str[:co.MaxChars] + fmt.Sprintf("...[TRUNCATED %d chars]", len(str)-co.MaxChars)
							// Marshal the truncated string to get valid JSON string representation (with quotes and escapes)
							replacement, _ := json.Marshal(truncated)
							mods = append(mods, modification{
								start:       text.Index,
								end:         text.Index + len(text.Raw),
								replacement: replacement,
							})
						}
					}
					return true // continue
				})
			}

			if len(mods) > 0 {
				// Rebuild the body with modifications
				var newBody bytes.Buffer
				// Pre-allocate buffer with rough size estimate (original size should be enough since we are shortening)
				newBody.Grow(len(bodyBytes))

				lastPos := 0
				for _, mod := range mods {
					// Append data before the modification
					newBody.Write(bodyBytes[lastPos:mod.start])
					// Append the replacement
					newBody.Write(mod.replacement)
					// Advance position
					lastPos = mod.end
				}
				// Append remaining data
				newBody.Write(bodyBytes[lastPos:])

				// Write new body
				if _, err := w.ResponseWriter.Write(newBody.Bytes()); err != nil {
					_ = err
				}
				return
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
	shouldBuffer bool
	checked      bool
}

func (w *responseBuffer) isBuffering() bool {
	// If explicitly set to false, then false.
	if w.checked && !w.shouldBuffer {
		return false
	}
	// If true or nil (default), we assume buffering for now, OR we haven't written yet.
	return true
}

func (w *responseBuffer) checkBuffer() {
	if !w.checked {
		ct := w.Header().Get("Content-Type")
		// We only buffer application/json.
		w.shouldBuffer = ct == "application/json" || strings.HasPrefix(ct, "application/json;")
		w.checked = true
	}
}

// Write writes the data to the buffer or the underlying ResponseWriter.
func (w *responseBuffer) Write(b []byte) (int, error) {
	w.checkBuffer()

	if w.shouldBuffer {
		return w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// WriteString writes the string data to the buffer or the underlying ResponseWriter.
func (w *responseBuffer) WriteString(s string) (int, error) {
	w.checkBuffer()

	if w.shouldBuffer {
		return w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}
