package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetEntropy(t *testing.T) {
	app := &Application{}
	handler := app.handleGetEntropy()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/entropy/{session_id}", handler).Methods("GET")

	req, err := http.NewRequest("GET", "/api/v1/entropy/test-session-123", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response EntropyScoreResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test-session-123", response.SessionID)
	assert.Equal(t, 0.85, response.CoherenceScore)
	assert.Equal(t, "aligned", response.Status)
}

func TestPostEntropyGate(t *testing.T) {
	app := &Application{}
	handler := app.handlePostEntropyGate()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/policy/entropy-gate", handler).Methods("POST")

	payload := []byte(`{"threshold": 0.7, "action": "revoke_write"}`)
	req, err := http.NewRequest("POST", "/api/v1/policy/entropy-gate", bytes.NewBuffer(payload))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)

	// Test invalid payload
	payloadInvalid := []byte(`{"threshold": 1.5, "action": "revoke_write"}`)
	reqInvalid, _ := http.NewRequest("POST", "/api/v1/policy/entropy-gate", bytes.NewBuffer(payloadInvalid))
	reqInvalid.Header.Set("Content-Type", "application/json")

	rrInvalid := httptest.NewRecorder()
	router.ServeHTTP(rrInvalid, reqInvalid)

	assert.Equal(t, http.StatusBadRequest, rrInvalid.Code)
}
