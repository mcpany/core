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
//
// maxChars is the maxChars.
//
// Returns the result.
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
//
// next is the next.
//
// Returns the result.
func (co *ContextOptimizer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wb := bufferPool.Get().(*responseBuffer)
		wb.ResponseWriter = w
		wb.body.Reset()
		wb.checked = false
		wb.shouldBuffer = false
		wb.status = http.StatusOK // Default status
		wb.wroteHeader = false

		defer func() {
			wb.ResponseWriter = nil // Avoid holding reference
			// If buffer is too large (>1MB), replace it with a new one to release memory
			if wb.body.Cap() > 1024*1024 {
				wb.body = &bytes.Buffer{}
			}
			bufferPool.Put(wb)
		}()

		next.ServeHTTP(wb, r)

		// If we didn't buffer, it means we passed through the original writer (streaming).
		// In that case, we don't need to do anything here.
		if !wb.isBuffering() {
			return
		}

		// Only check successful JSON responses
		// Note: w.shouldBuffer should already cover the content-type check, but we double check for safety
		contentType := wb.Header().Get("Content-Type")
		if wb.status == http.StatusOK && (contentType == "application/json" || strings.HasPrefix(contentType, "application/json;")) {
			bodyBytes := wb.body.Bytes()

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
				// Use original writer to write. We must write header first.
				// Update Content-Length because body size changed
				wb.Header().Set("Content-Length", fmt.Sprintf("%d", newBody.Len()))
				wb.ResponseWriter.WriteHeader(wb.status)
				if _, err := wb.ResponseWriter.Write(newBody.Bytes()); err != nil {
					_ = err
				}
				return
			}
		}

		// If not modified, write original body
		wb.ResponseWriter.WriteHeader(wb.status)
		if _, err := wb.ResponseWriter.Write(wb.body.Bytes()); err != nil {
			_ = err
		}
	})
}

type responseBuffer struct {
	http.ResponseWriter
	body         *bytes.Buffer
	shouldBuffer bool
	checked      bool
	status       int
	wroteHeader  bool
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
//
// b is the b.
//
// Returns the result.
// Returns an error if the operation fails.
func (w *responseBuffer) Write(b []byte) (int, error) {
	w.checkBuffer()

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.shouldBuffer {
		return w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// WriteHeader captures the status code and decides whether to buffer based on headers.
//
// statusCode is the HTTP status code to write.
func (w *responseBuffer) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.status = statusCode
	w.wroteHeader = true
	w.checkBuffer()
	if !w.shouldBuffer {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}
