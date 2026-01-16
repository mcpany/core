// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

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

var bufferPool = sync.Pool{
	New: func() any {
		return &responseBuffer{
			body: &bytes.Buffer{},
		}
	},
}

// Handler returns the middleware handler.
func (co *ContextOptimizer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := bufferPool.Get().(*responseBuffer)
		buf.reset(w)
		defer bufferPool.Put(buf)

		next.ServeHTTP(buf, r)

		// Process response
		co.processResponse(w, buf)
	})
}

func (co *ContextOptimizer) processResponse(w http.ResponseWriter, buf *responseBuffer) {
	// If statusCode was not set, it defaults to 200 (if body was written)
	if buf.statusCode == 0 {
		buf.statusCode = http.StatusOK
	}

	contentType := buf.Header().Get("Content-Type")
	shouldProcess := buf.statusCode == http.StatusOK &&
		(contentType == "application/json" || strings.HasPrefix(contentType, "application/json;"))

	if !shouldProcess {
		// Just copy everything
		if buf.statusCode != 0 {
			w.WriteHeader(buf.statusCode)
		}
		_, _ = w.Write(buf.body.Bytes())
		return
	}

	bodyBytes := buf.body.Bytes()

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
		newBody.Grow(len(bodyBytes))

		lastPos := 0
		for _, mod := range mods {
			newBody.Write(bodyBytes[lastPos:mod.start])
			newBody.Write(mod.replacement)
			lastPos = mod.end
		}
		newBody.Write(bodyBytes[lastPos:])

		// Update Content-Length if necessary (net/http handles it usually if we write)
		// w.Header().Set("Content-Length", strconv.Itoa(newBody.Len()))
		w.WriteHeader(buf.statusCode)
		_, _ = w.Write(newBody.Bytes())
		return
	}

	// No modifications
	w.WriteHeader(buf.statusCode)
	_, _ = w.Write(bodyBytes)
}

type responseBuffer struct {
	w          http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rb *responseBuffer) reset(w http.ResponseWriter) {
	rb.w = w
	rb.body.Reset()
	rb.statusCode = 0
}

func (rb *responseBuffer) Header() http.Header {
	return rb.w.Header()
}

func (rb *responseBuffer) Write(b []byte) (int, error) {
	return rb.body.Write(b)
}

func (rb *responseBuffer) WriteHeader(statusCode int) {
	rb.statusCode = statusCode
}
