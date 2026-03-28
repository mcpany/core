package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCMCSProviderMiddleware(t *testing.T) {
	policies := map[string][]string{
		"researcher": {"read", "summarize"},
		"coder":      {"write", "delete"},
		"admin":      {"*"},
	}
	middleware := NewCMCSProviderMiddleware(policies)

	tests := []struct {
		name       string
		headerVal  string
		wantStatus int
	}{
		{
			name:       "No Header",
			headerVal:  "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Authorized Researcher",
			headerVal:  "researcher:read",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Unauthorized Researcher",
			headerVal:  "researcher:delete",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Authorized Coder",
			headerVal:  "coder:write",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Admin Wildcard",
			headerVal:  "admin:nuke_db",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Unknown Role",
			headerVal:  "hacker:read",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Malformed Token",
			headerVal:  "researcher-read",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.headerVal != "" {
				req.Header.Set("X-Mesh-Token", tt.headerVal)
			}

			rr := httptest.NewRecorder()
			handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}
