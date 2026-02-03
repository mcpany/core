package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecureDefaults(t *testing.T) {
	// Case 1: No secret, no insecure mode -> Error
	_, err := NewHandler(Config{WebhookSecret: "", InsecureMode: false})
	if err == nil {
		t.Error("Expected error when starting without secret and not in insecure mode, got nil")
	} else if !strings.Contains(err.Error(), "WEBHOOK_SECRET is required") {
		t.Errorf("Expected error message to contain requirement info, got: %v", err)
	}

	// Case 2: No secret, insecure mode -> OK
	_, err = NewHandler(Config{WebhookSecret: "", InsecureMode: true})
	if err != nil {
		t.Errorf("Expected no error when insecure mode is enabled, got: %v", err)
	}
}

func TestAuthEnforcement(t *testing.T) {
	// standard-webhooks requires a base64 encoded secret
	secret := "dGVzdC1zZWNyZXQ=" // base64 for "test-secret"
	handler, err := NewHandler(Config{WebhookSecret: secret, InsecureMode: false})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Case 1: Request without signature -> 401
	req := httptest.NewRequest("POST", "/markdown", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for missing signature, got %d", w.Code)
	}

	// Case 2: Request with invalid signature -> 401
	req = httptest.NewRequest("POST", "/markdown", strings.NewReader("{}"))
	req.Header.Set("webhook-id", "msg_123")
	req.Header.Set("webhook-timestamp", "1234567890")
	req.Header.Set("webhook-signature", "v1,invalid-sig")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for invalid signature, got %d", w.Code)
	}
}

func TestDoSProtection(t *testing.T) {
	// Enable insecure mode to test body limit without needing signature
	handler, err := NewHandler(Config{WebhookSecret: "", InsecureMode: true})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Case 1: Body > 1MB -> 413 or 400
	// We create a large buffer
	largeBody := make([]byte, 1024*1024+100)
	req := httptest.NewRequest("POST", "/markdown", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// In Insecure Mode, the handler (MarkdownHandler) reads the body.
	// Since MaxBytesReader returns an error when limit is exceeded, the handler
	// catches it and returns 400 Bad Request (failed to parse CloudEvent).
	// If Auth was enabled, the middleware would catch it and return 413.
	// Both are acceptable as long as the read was limited.
	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 Payload Too Large or 400 Bad Request, got %d", w.Code)
	}
}
