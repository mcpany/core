package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAuthenticationInterceptor_AuthenticateRequest(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		headerKey     string
		headerValue   string
		expectError   bool
		expectedError error
	}{
		{
			name:        "Valid API Key",
			apiKey:      "test-key",
			headerKey:   APIKeyHeader,
			headerValue: "test-key",
			expectError: false,
		},
		{
			name:          "Missing API Key",
			apiKey:        "test-key",
			headerKey:     "Another-Header",
			headerValue:   "some-value",
			expectError:   true,
			expectedError: ErrMissingAPIKey,
		},
		{
			name:          "Invalid API Key",
			apiKey:        "test-key",
			headerKey:     APIKeyHeader,
			headerValue:   "invalid-key",
			expectError:   true,
			expectedError: ErrInvalidAPIKey,
		},
		{
			name:        "No API Key Configured",
			apiKey:      "",
			headerKey:   "any",
			headerValue: "any",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewAuthenticationInterceptor(tt.apiKey)
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(tt.headerKey, tt.headerValue)

			err := interceptor.AuthenticateRequest(req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthenticationInterceptor_Wrap(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		headerKey      string
		headerValue    string
		expectedStatus int
	}{
		{
			name:           "Valid API Key",
			apiKey:         "test-key",
			headerKey:      APIKeyHeader,
			headerValue:    "test-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing API Key",
			apiKey:         "test-key",
			headerKey:      "Another-Header",
			headerValue:    "some-value",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid API Key",
			apiKey:         "test-key",
			headerKey:      APIKeyHeader,
			headerValue:    "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "No API Key Configured",
			apiKey:         "",
			headerKey:      "any",
			headerValue:    "any",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewAuthenticationInterceptor(tt.apiKey)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			wrappedHandler := interceptor.Wrap(handler)

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(tt.headerKey, tt.headerValue)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestAuthenticationInterceptor_authenticate(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		ctx         context.Context
		expectError bool
	}{
		{
			name:   "Valid API Key",
			apiKey: "test-key",
			ctx:    metadata.NewIncomingContext(context.Background(), metadata.Pairs(APIKeyHeader, "test-key")),
		},
		{
			name:        "Missing API Key",
			apiKey:      "test-key",
			ctx:         context.Background(),
			expectError: true,
		},
		{
			name:        "Invalid API Key",
			apiKey:      "test-key",
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.Pairs(APIKeyHeader, "invalid-key")),
			expectError: true,
		},
		{
			name:   "No API Key Configured",
			apiKey: "",
			ctx:    context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewAuthenticationInterceptor(tt.apiKey)
			err := interceptor.authenticate(tt.ctx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
