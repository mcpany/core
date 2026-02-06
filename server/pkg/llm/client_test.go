package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	testCases := []struct {
		name           string
		apiKey         string
		req            ChatRequest
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedResp   *ChatResponse
		expectedErr    string
	}{
		{
			name:   "Success",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)
				assert.Equal(t, "Hello", reqBody.Messages[0].Content)

				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{
						{
							Message: struct {
								Content string `json:"content"`
							}{
								Content: "Hi there!",
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hi there!",
			},
		},
		{
			name:   "API Error",
			apiKey: "test-key",
			req:    ChatRequest{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API Key"}}`))
			},
			expectedErr: "openai api error (status 401)",
		},
		{
			name:   "OpenAI Logic Error",
			apiKey: "test-key",
			req:    ChatRequest{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Something went wrong",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "openai error: Something went wrong",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-key",
			req:    ChatRequest{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "no choices returned",
		},
		{
			name:   "Malformed JSON",
			apiKey: "test-key",
			req:    ChatRequest{},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{invalid json`))
			},
			expectedErr: "failed to decode response",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.mockHandler))
			defer server.Close()

			client := NewOpenAIClient(tc.apiKey, server.URL)
			resp, err := client.ChatCompletion(context.Background(), tc.req)

			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestOpenAIClient_NetworkError(t *testing.T) {
	// Create a client pointing to a closed port
	// We use a reserved invalid port to ensure connection refusal
	client := NewOpenAIClient("test-key", "http://127.0.0.1:0")
	_, err := client.ChatCompletion(context.Background(), ChatRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}
