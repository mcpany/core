package integration

import (
	"net/http"
)

// APIKeyInjectingTransport is an http.RoundTripper that injects an X-API-Key header.
type APIKeyInjectingTransport struct {
	Base   http.RoundTripper
	APIKey string
}

// RoundTrip executes a single HTTP transaction.
func (t *APIKeyInjectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.APIKey != "" {
		req.Header.Set("X-API-Key", t.APIKey)
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// NewAPIKeyClient returns an http.Client that injects the given API key.
func NewAPIKeyClient(apiKey string) *http.Client {
	return &http.Client{
		Transport: &APIKeyInjectingTransport{APIKey: apiKey},
	}
}
