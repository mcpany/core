package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/alerts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type mockAlertsManager struct {
	createdAlerts []*alerts.Alert
	updatedAlerts []*alerts.Alert
}

func (m *mockAlertsManager) ListAlerts() []*alerts.Alert { return nil }
func (m *mockAlertsManager) GetAlert(id string) *alerts.Alert { return nil }
func (m *mockAlertsManager) CreateAlert(alert *alerts.Alert) *alerts.Alert {
	alert.ID = "AL-TEST-1"
	m.createdAlerts = append(m.createdAlerts, alert)
	return alert
}
func (m *mockAlertsManager) UpdateAlert(id string, alert *alerts.Alert) *alerts.Alert {
	m.updatedAlerts = append(m.updatedAlerts, alert)
	return alert
}
func (m *mockAlertsManager) ListRules() []*alerts.AlertRule { return nil }
func (m *mockAlertsManager) GetRule(id string) *alerts.AlertRule { return nil }
func (m *mockAlertsManager) CreateRule(rule *alerts.AlertRule) *alerts.AlertRule { return nil }
func (m *mockAlertsManager) UpdateRule(id string, rule *alerts.AlertRule) *alerts.AlertRule { return nil }
func (m *mockAlertsManager) DeleteRule(id string) error { return nil }

func TestAlertsIntegration(t *testing.T) {
	mockManager := &mockAlertsManager{}
	SetGlobalAlertsManager(mockManager)

    // Reset active alerts
    activeAlertsMu.Lock()
    activeAlerts = make(map[string]string)
    activeAlertsMu.Unlock()

	// 1. Setup a failing service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	addr := server.Listener.Addr().String()
	upstreamConfig := configv1.UpstreamServiceConfig_builder{
		Name: lo.ToPtr("test-alert-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: &addr,
			HealthCheck: configv1.HttpHealthCheck_builder{
				Url:          lo.ToPtr(server.URL),
				ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
			}.Build(),
		}.Build(),
	}.Build()

	checker := NewChecker(upstreamConfig)
	assert.NotNil(t, checker)

	// 2. Trigger check (Failing)
	_ = checker.Check(context.Background())

	// Verify Alert Created
	if assert.NotEmpty(t, mockManager.createdAlerts) {
		assert.Equal(t, "test-alert-service", mockManager.createdAlerts[0].Service)
		assert.Equal(t, alerts.SeverityCritical, mockManager.createdAlerts[0].Severity)
		assert.Equal(t, alerts.StatusActive, mockManager.createdAlerts[0].Status)
	}
}
