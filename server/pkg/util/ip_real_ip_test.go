package util

import (
	"net/http"
	"testing"
)

func TestGetClientIP_XRealIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		trustProxy bool
		want       string
	}{
		{
			name:       "No headers, no trust",
			remoteAddr: "1.2.3.4:1234",
			headers:    nil,
			trustProxy: false,
			want:       "1.2.3.4",
		},
		{
			name:       "X-Real-IP present, trust",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Real-IP": "5.6.7.8",
			},
			trustProxy: true,
			want:       "5.6.7.8",
		},
		{
			name:       "X-Real-IP present, no trust",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Real-IP": "5.6.7.8",
			},
			trustProxy: false,
			want:       "1.2.3.4",
		},
		{
			name:       "X-Real-IP and XFF, trust",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Real-IP":       "5.6.7.8",
				"X-Forwarded-For": "9.9.9.9, 5.6.7.8",
			},
			trustProxy: true,
			want:       "5.6.7.8",
		},
		{
			name:       "Only XFF, trust",
			remoteAddr: "1.2.3.4:1234",
			headers: map[string]string{
				"X-Forwarded-For": "9.9.9.9, 5.6.7.8",
			},
			trustProxy: true,
			want:       "9.9.9.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if got := GetClientIP(req, tt.trustProxy); got != tt.want {
				t.Errorf("GetClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
