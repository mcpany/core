// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProposeHandler_Valid(t *testing.T) {
	hub := NewA2AMessagingHub(nil)
	payload := map[string]interface{}{
		"intent": "write_file",
		"data":   "hello",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/a2a/propose", bytes.NewReader(body))
	w := httptest.NewRecorder()

	hub.ProposeHandler(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestProposeHandler_MissingIntent(t *testing.T) {
	hub := NewA2AMessagingHub(nil)
	payload := map[string]interface{}{
		"data": "hello",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/a2a/propose", bytes.NewReader(body))
	w := httptest.NewRecorder()

	hub.ProposeHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMailboxHandler_MissingAgentID(t *testing.T) {
	hub := NewA2AMessagingHub(nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/a2a/mailbox", nil)
	w := httptest.NewRecorder()

	hub.MailboxHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
