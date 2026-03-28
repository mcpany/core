package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPSCMiddleware(t *testing.T) {
	middleware := NewIPSCMiddleware(3) // Max 3 cycles

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
			name:       "Below Threshold",
			headerVal:  "cycle=1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "At Threshold",
			headerVal:  "cycle=3",
			wantStatus: http.StatusTooManyRequests,
		},
		{
			name:       "Above Threshold",
			headerVal:  "cycle=5",
			wantStatus: http.StatusTooManyRequests,
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
				req.Header.Set("X-UACO-IPSC", tt.headerVal)
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
