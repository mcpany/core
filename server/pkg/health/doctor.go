// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckResult represents a single check result.
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// DoctorReport represents the full doctor report.
type DoctorReport struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Doctor is the health check handler.
type Doctor struct {
	// Add dependencies here if needed, e.g. db connection
}

// NewDoctor creates a new Doctor.
func NewDoctor() *Doctor {
	return &Doctor{}
}

// Handler returns the gin handler.
func (d *Doctor) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		report := DoctorReport{
			Status:    "healthy",
			Timestamp: time.Now(),
			Checks:    make(map[string]CheckResult),
		}

		// Check Internet
		start := time.Now()
		req, err := http.NewRequestWithContext(c.Request.Context(), "GET", "https://www.google.com", nil)
		if err == nil {
			var resp *http.Response
			resp, err = http.DefaultClient.Do(req)
			if resp != nil {
				defer func() {
					_ = resp.Body.Close()
				}()
			}
		}

		if err != nil {
			report.Checks["internet"] = CheckResult{
				Status:  "degraded",
				Message: err.Error(),
			}
			report.Status = "degraded"
		} else {
			report.Checks["internet"] = CheckResult{
				Status:  "ok",
				Latency: time.Since(start).String(),
			}
		}

		// We could add Redis/DB checks here if we had the connections injected.
		// For MVP, just internet connectivity is a good "Doctor" check for "Why can't I access my tools?"

		c.JSON(http.StatusOK, report)
	}
}
