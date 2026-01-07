// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/health/doctor", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var report DiagnosticReport
	err = json.Unmarshal(rr.Body.Bytes(), &report)
	if err != nil {
		t.Errorf("handler returned invalid json: %v", err)
	}

	if report.Status != "ok" {
		t.Errorf("handler returned wrong status: got %v want %v",
			report.Status, "ok")
	}

	if _, ok := report.Environment["GOOS"]; !ok {
		t.Errorf("handler missing Environment.GOOS")
	}
}
