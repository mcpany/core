package benchmark

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

type ExecutionRequest struct {
	ToolName   string
	ToolInputs []byte
}

func getCacheKeyOriginal(req *ExecutionRequest) string {
	// Normalize ToolInputs if they are JSON
	var normalizedInputs []byte
	if len(req.ToolInputs) > 0 {
		// Optimization: Check if it looks like a JSON object or array before unmarshaling
		// Skip leading whitespace (simplified check)
		var firstChar byte
		for _, b := range req.ToolInputs {
			if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
				firstChar = b
				break
			}
		}

		if firstChar == '{' || firstChar == '[' {
			var input any
			// We use standard json.Unmarshal which sorts map keys when Marshaling back.
			if err := json.Unmarshal(req.ToolInputs, &input); err == nil {
				if marshaled, err := json.Marshal(input); err == nil {
					normalizedInputs = marshaled
				}
			}
		}
	}

	// Fallback to raw bytes if unmarshal fails or empty
	if normalizedInputs == nil {
		normalizedInputs = req.ToolInputs
	}

	// Optimization: Hash the normalized inputs to keep the cache key short and fixed length.
	// This avoids using potentially large JSON strings as map keys.
	hash := sha256.Sum256(normalizedInputs)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s:%s", req.ToolName, hashStr)
}

func getCacheKeyOptimized(req *ExecutionRequest) string {
	// Normalize ToolInputs if they are JSON
	var normalizedInputs []byte
	if len(req.ToolInputs) > 0 {
		// Check if it looks like a JSON object or array before unmarshaling
		// Skip leading whitespace (simplified check)
		var firstChar byte
		idx := 0
		for i, b := range req.ToolInputs {
			if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
				firstChar = b
				idx = i
				break
			}
		}

		if firstChar == '{' || firstChar == '[' {
			// Optimization: If the JSON is very short, unmarshal/marshal overhead dominates.
            // But we can't easily skip it without risking cache misses.

            // Try to use RawMessage to defer parsing? No, we need value to sort keys.

			var input any
			// We use standard json.Unmarshal which sorts map keys when Marshaling back.
			if err := json.Unmarshal(req.ToolInputs[idx:], &input); err == nil {
				if marshaled, err := json.Marshal(input); err == nil {
					normalizedInputs = marshaled
				}
			}
		}
	}

	// Fallback to raw bytes if unmarshal fails or empty
	if normalizedInputs == nil {
		normalizedInputs = req.ToolInputs
	}

	hash := sha256.Sum256(normalizedInputs)
	hashStr := hex.EncodeToString(hash[:])

	return req.ToolName + ":" + hashStr
}

func BenchmarkGetCacheKeyOriginal(b *testing.B) {
    req := &ExecutionRequest{
        ToolName: "test-tool",
        ToolInputs: []byte(`{"b": 2, "a": 1, "c": [1, 2, 3], "d": {"x": "y"}}`),
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        getCacheKeyOriginal(req)
    }
}

func BenchmarkGetCacheKeyOptimized(b *testing.B) {
    req := &ExecutionRequest{
        ToolName: "test-tool",
        ToolInputs: []byte(`{"b": 2, "a": 1, "c": [1, 2, 3], "d": {"x": "y"}}`),
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        getCacheKeyOptimized(req)
    }
}
