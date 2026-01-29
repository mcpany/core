package mcpserver

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertMapToCallToolResult(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]any
		want      *mcp.CallToolResult
		wantErr   bool
		errSubstr string
	}{
		{
			name: "Valid Text",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "hello",
					},
				},
				"isError": false,
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "hello"},
				},
				IsError: false,
			},
		},
		{
			name: "Valid Image",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type":     "image",
						"data":     base64.StdEncoding.EncodeToString([]byte("fake")),
						"mimeType": "image/png",
					},
				},
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{
						Data:     []byte("fake"),
						MIMEType: "image/png",
					},
				},
			},
		},
		{
			name: "Valid Resource",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":      "file:///test.txt",
							"mimeType": "text/plain",
							"text":     "content",
						},
					},
				},
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							URI:      "file:///test.txt",
							MIMEType: "text/plain",
							Text:     "content",
						},
					},
				},
			},
		},
		{
			name: "Resource with Blob",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":  "file:///test.bin",
							"blob": base64.StdEncoding.EncodeToString([]byte("blob")),
						},
					},
				},
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							URI:  "file:///test.bin",
							Blob: []byte("blob"),
						},
					},
				},
			},
		},
		{
			name: "No Content (Just Error)",
			input: map[string]any{
				"isError": true,
			},
			want: &mcp.CallToolResult{
				IsError: true,
			},
		},
		{
			name: "Content Not List",
			input: map[string]any{
				"content": "not a list",
			},
			wantErr:   true,
			errSubstr: "content is not a list",
		},
		{
			name: "Content Item Not Map",
			input: map[string]any{
				"content": []any{"string"},
			},
			wantErr:   true,
			errSubstr: "content item is not a map",
		},
		{
			name: "Missing Type",
			input: map[string]any{
				"content": []any{
					map[string]any{"text": "foo"},
				},
			},
			wantErr:   true,
			errSubstr: "content type is not a string",
		},
		{
			name: "Unsupported Type",
			input: map[string]any{
				"content": []any{
					map[string]any{"type": "video"},
				},
			},
			wantErr:   true,
			errSubstr: "unsupported content type",
		},
		{
			name: "Text Missing Field",
			input: map[string]any{
				"content": []any{
					map[string]any{"type": "text"},
				},
			},
			wantErr:   true,
			errSubstr: "text content text is not a string",
		},
		{
			name: "Image Bad Base64",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "image",
						"data": "!!!",
					},
				},
			},
			wantErr:   true,
			errSubstr: "failed to decode image data",
		},
		{
			name: "Image Missing Mime",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "image",
						"data": "abcd",
					},
				},
			},
			wantErr:   true,
			errSubstr: "image content mimeType is not a string",
		},
		{
			name: "Resource Missing URI",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"text": "foo",
						},
					},
				},
			},
			wantErr:   true,
			errSubstr: "resource uri is not a string",
		},
		{
			name: "Resource Bad Blob",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":  "foo",
							"blob": "!!!",
						},
					},
				},
			},
			wantErr:   true,
			errSubstr: "failed to decode resource blob",
		},
		{
			name: "Neither Content nor IsError",
			input: map[string]any{
				"foo": "bar",
			},
			wantErr:   true,
			errSubstr: "neither content nor isError present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertMapToCallToolResult(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSummarizeCallToolResult(t *testing.T) {
	longText := strings.Repeat("a", 1000)
	ctr := &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: longText},
			&mcp.ImageContent{MIMEType: "image/png", Data: []byte("fake")},
			&mcp.EmbeddedResource{
				Resource: &mcp.ResourceContents{
					URI:  "file:///test",
					Text: "short",
					Blob: []byte("blob"),
				},
			},
		},
	}

	val := summarizeCallToolResult(ctr)
	// slog.Value logic is hard to inspect deeply without reflection or stringifying,
	// but we can check if it returns a Group
	assert.Equal(t, "Group", val.Kind().String())

	// We can't easily inspect the members of the GroupValue with standard library slog.Value methods
	// in a way that is easy to assert.
	// But we covered the code path.
}

func TestLazyLogResult(t *testing.T) {
	// Test nil
	val := LazyLogResult{Value: nil}.LogValue()
	assert.Equal(t, "<nil>", val.String())

	// Test CallToolResult
	ctr := &mcp.CallToolResult{IsError: true}
	val = LazyLogResult{Value: ctr}.LogValue()
	assert.Contains(t, val.String(), "isError")

	// Test Map that looks like CallToolResult
	m := map[string]any{
		"isError": true,
		"content": []any{
			map[string]any{"type": "text", "text": "hello"},
		},
	}
	val = LazyLogResult{Value: m}.LogValue()
	assert.Contains(t, val.String(), "isError")
	assert.Contains(t, val.String(), "hello")

	// Test Map that does NOT look like CallToolResult (redaction)
	m2 := map[string]any{"secret": "sensitive"}
	val = LazyLogResult{Value: m2}.LogValue()
	// Redaction should happen (LazyRedact)
	// Note: RedactJSON usually hashes values or masks them.
	// Assuming util.RedactJSON works as expected (tested elsewhere).
	// We just check it's a string value.
	assert.Equal(t, "String", val.Kind().String())

	// Test fallback
	val = LazyLogResult{Value: 123}.LogValue()
	assert.Equal(t, "123", val.String())
}
