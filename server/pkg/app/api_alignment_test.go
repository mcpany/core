// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleActiveIntentAlignment(t *testing.T) {
	app := &Application{}

	handler := app.handleActiveIntentAlignment()

	req, err := http.NewRequest("GET", "/api/v1/alignment/status", nil)
	require.NoError(t, err)

	// We run it multiple times to ensure the entropy bounds (0-100) are hit and logic is exercised
	for i := 0; i < 50; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var statuses []SubagentStatus
		err = json.Unmarshal(rr.Body.Bytes(), &statuses)
		require.NoError(t, err)

		// There are initially 4 subagents defined
		assert.Len(t, statuses, 4)

		for _, status := range statuses {
			assert.NotEmpty(t, status.ID)
			assert.NotEmpty(t, status.Name)

			assert.GreaterOrEqual(t, status.EntropyScore, 0.0)
			assert.LessOrEqual(t, status.EntropyScore, 100.0)

			if status.EntropyScore > 85 {
				assert.Equal(t, "hijacked", status.Status)
			} else if status.EntropyScore > 50 {
				assert.Equal(t, "drifting", status.Status)
			} else {
				assert.Equal(t, "aligned", status.Status)
			}

			assert.NotEqual(t, int64(0), status.LastHeartbeat)
		}
	}
}
