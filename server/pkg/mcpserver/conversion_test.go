// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"encoding/base64"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertMapToCallToolResult_Comprehensive(t *testing.T) {
	// Helper to encode string to base64
	b64 := func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		name      string
		input     map[string]any
		want      *mcp.CallToolResult
		wantErr   bool
		errSubstr string
	}{
		{
			name: "Success_TextContent",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "Hello World",
					},
				},
				"isError": false,
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello World"},
				},
				IsError: false,
			},
		},
		{
			name: "Success_ImageContent",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type":     "image",
						"data":     b64("image data"),
						"mimeType": "image/png",
					},
				},
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{
						Data:     []byte("image data"),
						MIMEType: "image/png",
					},
				},
				IsError: false,
			},
		},
		{
			name: "Success_ResourceContent",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":      "file:///test.txt",
							"mimeType": "text/plain",
							"text":     "resource text",
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
							Text:     "resource text",
						},
					},
				},
				IsError: false,
			},
		},
		{
			name: "Success_ResourceContent_Blob",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":  "file:///test.bin",
							"blob": b64("blob data"),
						},
					},
				},
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							URI:  "file:///test.bin",
							Blob: []byte("blob data"),
						},
					},
				},
				IsError: false,
			},
		},
		{
			name: "Success_ResourceContent_BlobBytes",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":  "file:///test.bin",
							"blob": []byte("blob bytes"),
						},
					},
				},
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							URI:  "file:///test.bin",
							Blob: []byte("blob bytes"),
						},
					},
				},
				IsError: false,
			},
		},
		{
			name: "Success_MixedContent",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "Part 1",
					},
					map[string]any{
						"type": "text",
						"text": "Part 2",
					},
				},
				"isError": true,
			},
			want: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Part 1"},
					&mcp.TextContent{Text: "Part 2"},
				},
				IsError: true,
			},
		},
		{
			name: "Success_ErrorOnly",
			input: map[string]any{
				"isError": true,
			},
			want: &mcp.CallToolResult{
				IsError: true,
			},
		},
		{
			name:      "Error_MissingContentAndIsError",
			input:     map[string]any{},
			wantErr:   true,
			errSubstr: "neither content nor isError present",
		},
		{
			name: "Error_ContentNotList",
			input: map[string]any{
				"content": "not a list",
			},
			wantErr:   true,
			errSubstr: "content is not a list",
		},
		{
			name: "Error_ContentItemNotMap",
			input: map[string]any{
				"content": []any{
					"not a map",
				},
			},
			wantErr:   true,
			errSubstr: "content item is not a map",
		},
		{
			name: "Error_MissingType",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"text": "missing type",
					},
				},
			},
			wantErr:   true,
			errSubstr: "content type is not a string",
		},
		{
			name: "Error_InvalidType",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": 123,
					},
				},
			},
			wantErr:   true,
			errSubstr: "content type is not a string",
		},
		{
			name: "Error_UnsupportedType",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "audio",
					},
				},
			},
			wantErr:   true,
			errSubstr: "unsupported content type",
		},
		{
			name: "Error_Text_InvalidTextField",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "text",
						"text": 123,
					},
				},
			},
			wantErr:   true,
			errSubstr: "text content text is not a string",
		},
		{
			name: "Error_Image_MissingData",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type":     "image",
						"mimeType": "image/png",
					},
				},
			},
			wantErr:   true,
			errSubstr: "image content data is not a string",
		},
		{
			name: "Error_Image_InvalidBase64",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type":     "image",
						"data":     "invalid-base64",
						"mimeType": "image/png",
					},
				},
			},
			wantErr:   true,
			errSubstr: "failed to decode image data",
		},
		{
			name: "Error_Image_MissingMimeType",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "image",
						"data": b64("data"),
					},
				},
			},
			wantErr:   true,
			errSubstr: "image content mimeType is not a string",
		},
		{
			name: "Error_Resource_MissingResourceMap",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
					},
				},
			},
			wantErr:   true,
			errSubstr: "resource content resource is not a map",
		},
		{
			name: "Error_Resource_MissingURI",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type":     "resource",
						"resource": map[string]any{},
					},
				},
			},
			wantErr:   true,
			errSubstr: "resource uri is not a string",
		},
		{
			name: "Error_Resource_InvalidBlobBase64",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":  "file:///test",
							"blob": "invalid-base64",
						},
					},
				},
			},
			wantErr:   true,
			errSubstr: "failed to decode resource blob",
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
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
