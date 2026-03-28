package app

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/middleware"
)

type HITLState struct {
	mu        sync.RWMutex
	approvals map[string]middleware.HITLApprovalRequest
}

func newHITLState() *HITLState {
	return &HITLState{
		approvals: make(map[string]middleware.HITLApprovalRequest),
	}
}

var globalHITLState = newHITLState()

func init() {
	// Seed some initial data for testing/UI purposes
	globalHITLState.mu.Lock()
	globalHITLState.approvals["1"] = middleware.HITLApprovalRequest{
		ExecutionID: "1",
		ToolName:    "database.drop_table",
		RequireMFA:  true,
	}
	globalHITLState.approvals["2"] = middleware.HITLApprovalRequest{
		ExecutionID: "2",
		ToolName:    "aws.terminate_instance",
		RequireMFA:  false,
	}
	globalHITLState.mu.Unlock()
}

func (a *Application) mountHITL(mux *http.ServeMux) {
	// First, subscribe to hitl.requests
	reqBus, err := bus.GetBus[middleware.HITLApprovalRequest](a.bus, "hitl.requests")
	if err == nil {
		reqBus.Subscribe(context.Background(), "hitl.requests", func(req middleware.HITLApprovalRequest) {
			globalHITLState.mu.Lock()
			globalHITLState.approvals[req.ExecutionID] = req
			globalHITLState.mu.Unlock()

			// Auto cleanup after 5 mins? This is just basic.
			go func(id string) {
				time.Sleep(5 * time.Minute)
				globalHITLState.mu.Lock()
				delete(globalHITLState.approvals, id)
				globalHITLState.mu.Unlock()
			}(req.ExecutionID)
		})
	} else {
		logging.GetLogger().Warn("Could not subscribe to hitl.requests", "error", err)
	}

	mux.HandleFunc("/hitl/approvals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		globalHITLState.mu.RLock()
		defer globalHITLState.mu.RUnlock()

		type uiApproval struct {
			ID         string `json:"id"`
			Tool       string `json:"tool"`
			Intent     string `json:"intent"`
			Status     string `json:"status"`
			RequireMfa bool   `json:"requireMfa"`
		}

		list := []uiApproval{}
		for _, req := range globalHITLState.approvals {
			list = append(list, uiApproval{
				ID:         req.ExecutionID,
				Tool:       req.ToolName,
				Intent:     "Pending verification for sensitive tool",
				Status:     "pending",
				RequireMfa: req.RequireMFA,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})

	mux.HandleFunc("/hitl/approvals/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID from /hitl/approvals/{id}
		id := r.URL.Path[len("/hitl/approvals/"):]

		var reqBody struct {
			Action  string `json:"action"`
			MfaCode string `json:"mfaCode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// In a real app we'd verify the MFA code here against the user profile

		// Publish the response
		resBus, err := bus.GetBus[middleware.HITLApprovalResponse](a.bus, "hitl.responses."+id)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		approved := reqBody.Action == "approved"

		err = resBus.Publish(r.Context(), "hitl.responses."+id, middleware.HITLApprovalResponse{
			ExecutionID: id,
			Approved:    approved,
		})

		if err != nil {
			http.Error(w, "Failed to publish response", http.StatusInternalServerError)
			return
		}

		// Remove from pending
		globalHITLState.mu.Lock()
		delete(globalHITLState.approvals, id)
		globalHITLState.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})
}
