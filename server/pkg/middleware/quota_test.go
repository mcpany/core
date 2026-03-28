package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQuotaMonitorMiddleware(t *testing.T) {
	middleware := NewQuotaMonitorMiddleware(1000)

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
			name:       "Below Quota",
			headerVal:  "500",
			wantStatus: http.StatusOK,
		},
		{
			name:       "At Quota",
			headerVal:  "1000",
			wantStatus: http.StatusPaymentRequired,
		},
		{
			name:       "Above Quota",
			headerVal:  "1500",
			wantStatus: http.StatusPaymentRequired,
		},
		{
			name:       "Malformed Header",
			headerVal:  "invalid",
			wantStatus: http.StatusOK, // Ignore malformed and pass through
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.headerVal != "" {
				req.Header.Set("X-Usage-Tokens", tt.headerVal)
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
