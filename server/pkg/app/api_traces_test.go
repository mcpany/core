package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func TestBulkDeleteTraces(t *testing.T) {
	app, _ := NewTestApp(t)

	// Add mock trace using the generate mock function already available on `a.handleDebugSeedTraces` or direct
	// We just verify the API responds properly.
	reqBody, _ := json.Marshal(map[string]interface{}{
		"ids": []string{"trace-to-delete"},
	})

	req, _ := http.NewRequest(http.MethodPost, "/traces/bulk-delete", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	app.apiHandler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}
