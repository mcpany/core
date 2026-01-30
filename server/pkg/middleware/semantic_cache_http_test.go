// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewHTTPEmbeddingProvider(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		bodyTemplateStr string
		wantErr         bool
	}{
		{
			name:            "valid",
			url:             "http://example.com",
			bodyTemplateStr: `{"input": "{{.input}}"}`,
			wantErr:         false,
		},
		{
			name:            "empty url",
			url:             "",
			bodyTemplateStr: `{"input": "{{.input}}"}`,
			wantErr:         true,
		},
		{
			name:            "invalid template",
			url:             "http://example.com",
			bodyTemplateStr: `{{.unclosed`,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHTTPEmbeddingProvider(tt.url, nil, tt.bodyTemplateStr, "$.data")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHTTPEmbeddingProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPEmbeddingProvider_Embed(t *testing.T) {
	tests := []struct {
		name             string
		headers          map[string]string
		bodyTemplateStr  string
		responseJSONPath string
		input            string
		mockHandler      func(w http.ResponseWriter, r *http.Request)
		want             []float32
		wantErr          bool
		checkReq         func(r *http.Request) error // optional hook to check request
	}{
		{
			name:             "happy path",
			bodyTemplateStr:  `{"prompt": "{{.input}}"}`,
			responseJSONPath: "$.embedding",
			input:            "hello world",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"embedding": [0.1, 0.2, 0.3]}`))
			},
			want:    []float32{0.1, 0.2, 0.3},
			wantErr: false,
		},
		{
			name:             "headers verification",
			headers:          map[string]string{"X-Auth": "secret"},
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Auth") != "secret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": [1.0]}`))
			},
			want:    []float32{1.0},
			wantErr: false,
		},
		{
			name:             "template execution failure - not easily reachable with string map",
			bodyTemplateStr:  `{{.missing}}`, // missing key usually just prints <no value> in Go templates unless Option("missingkey=error") is set.
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": [0.0]}`))
			},
			want:    []float32{0.0}, // It should actually succeed with empty string
			wantErr: false,
		},
		{
			name:             "upstream 500 error",
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`internal error`))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:             "malformed json response",
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{not valid json`))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:             "jsonpath extraction failure - path not found",
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.missing",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": [1.0]}`))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:             "jsonpath result not array",
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": "not an array"}`))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:             "jsonpath array contains non-numbers",
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": ["string", 1.0]}`))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:             "empty embedding",
			bodyTemplateStr:  `{}`,
			responseJSONPath: "$.data",
			input:            "test",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": []}`))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:             "request body verification",
			bodyTemplateStr:  `{"prompt": "{{.input}}"}`,
			responseJSONPath: "$.data",
			input:            "verify me",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if string(body) != `{"prompt": "verify me"}` {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data": [1.0]}`))
			},
			want:    []float32{1.0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			p, err := NewHTTPEmbeddingProvider(server.URL, tt.headers, tt.bodyTemplateStr, tt.responseJSONPath)
			if err != nil {
				t.Fatalf("NewHTTPEmbeddingProvider() error = %v", err)
			}
			// Override client to use the test server's client (though httptest URL is usually enough, this ensures TLS etc works if needed)
			// Actually NewHTTPEmbeddingProvider creates its own client.
			// But since we use server.URL, it should be fine as it's http.

			got, err := p.Embed(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Embed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Embed() got = %v, want %v", got, tt.want)
			}
		})
	}
}
