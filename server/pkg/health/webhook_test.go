package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendWebhook(t *testing.T) {
	// 1. Setup a mock webhook receiver server
	receivedPayloads := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Verify Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		receivedPayloads <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 2. Setup Config
	alertConfig := configv1.AlertConfig_builder{
		Enabled:    lo.ToPtr(true),
		WebhookUrl: lo.ToPtr(server.URL),
	}.Build()

	// 3. Set Global Config
	SetGlobalAlertConfig(alertConfig)

	// 4. Create a Checker for a service that we can toggle health
	// We'll use a Command Line check that passes initially
	upstreamConfig := configv1.UpstreamServiceConfig_builder{
		Name: lo.ToPtr("webhook-test-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: lo.ToPtr("true"), // Always succeeds
		}.Build(),
	}.Build()

	// 5. Create Checker
	checker := NewChecker(upstreamConfig)
	require.NotNil(t, checker)

	// 6. Run Check - Status should be UP.
	// Since status listeners are async or run during check, we need to trigger it.
	// NewChecker returns a health.Checker. Calling Check() triggers listeners.
	result := checker.Check(context.Background())
	assert.Equal(t, health.StatusUp, result.Status)

	// 7. Verify Webhook Received (Status Changed from Unknown -> Up)
	select {
	case payload := <-receivedPayloads:
		assert.Equal(t, "health_status_changed", payload["event"])
		assert.Equal(t, "webhook-test-service", payload["service"])
		assert.Equal(t, "up", payload["status"])
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for webhook")
	}

	// 8. Run Check Again - Status Unchanged (UP -> UP)
	// Deduplication should prevent another webhook
	result = checker.Check(context.Background())
	assert.Equal(t, health.StatusUp, result.Status)

	select {
	case <-receivedPayloads:
		t.Fatal("unexpected webhook received for unchanged status")
	case <-time.After(500 * time.Millisecond):
		// OK
	}

	// 9. Force Status Change (UP -> DOWN)
	// We can't easily change the command of an existing checker, but we can simulate
	// failure by making the check fail?
	// Actually, `commandLineCheck` uses `Execute`. `true` always succeeds.
	// To test state change, we might need a custom check function or a service we can control.
	// Let's use an HTTP service pointing to a server we can close.

	// New Test Case for State Change
	t.Run("StateChange", func(t *testing.T) {
		// Control Server
		var statusCode int = http.StatusOK
		controlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		}))
		defer controlServer.Close()

		addr := controlServer.Listener.Addr().String()
		svcConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("flapping-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &addr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          lo.ToPtr(controlServer.URL),
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(svcConfig)

		// Initial Check (UP)
		checker.Check(context.Background())
		// Consume initial webhook
		select {
		case <-receivedPayloads:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for initial webhook")
		}

		// Change to Down
		statusCode = http.StatusInternalServerError
		checker.Check(context.Background())

		// Verify Webhook (DOWN)
		// We might need a small loop because the checker might cache results or debounce?
		// NewChecker has cache duration of 1s in health.go options.
		// health.WithCacheDuration(1 * time.Second)
		// So we must wait > 1s before checking again to get fresh result.
		time.Sleep(1100 * time.Millisecond)
		checker.Check(context.Background())

		select {
		case payload := <-receivedPayloads:
			assert.Equal(t, "flapping-service", payload["service"])
			assert.Equal(t, "down", payload["status"])
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for down webhook")
		}
	})
}
