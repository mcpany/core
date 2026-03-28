package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExfiltrationTransportGateway(t *testing.T) {
	allowedDomains := []string{"api.anthropic.com", "*.openai.com"}
	gateway := NewExfiltrationTransportGateway(allowedDomains)

	tests := []struct {
		name       string
		host       string
		wantStatus int
	}{
		{
			name:       "Allowed Exact Match",
			host:       "api.anthropic.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Allowed Wildcard Match",
			host:       "api.openai.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Allowed Localhost",
			host:       "localhost",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Allowed IPv6 Localhost",
			host:       "[::1]:8080",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Blocked Domain",
			host:       "evil.com",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Blocked Subdomain Exact",
			host:       "sub.api.anthropic.com",
			wantStatus: http.StatusForbidden, // Exact match only for anthropic
		},
		{
			name:       "Allowed Subdomain Wildcard",
			host:       "sub.api.openai.com",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Host = tt.host

			rr := httptest.NewRecorder()
			handler := gateway.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}
		})
	}
}
