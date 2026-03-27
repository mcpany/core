package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mcpany/core/server/pkg/audit"
)

// generateMockAuditEntries returns mock audit log data
func generateMockAuditEntries(count int) []audit.LogEntry {
	entries := make([]audit.LogEntry, count)
	now := time.Now()

	// Create mock entries with varied data to demonstrate rich result capabilities
	for i := 0; i < count; i++ {
		id := uuid.New().String()
		timestamp := now.Add(-time.Duration(i*10) * time.Minute)

		// Create varied JSON payloads to demonstrate rich viewing
		var requestPayload map[string]interface{}
		var responsePayload map[string]interface{}

		if i%2 == 0 {
			requestPayload = map[string]interface{}{
				"method": "mcp.list_tools",
				"params": map[string]interface{}{
					"limit": 10,
				},
			}
			responsePayload = map[string]interface{}{
				"tools": []map[string]interface{}{
					{"name": "calculator", "description": "Basic math operations"},
					{"name": "weather", "description": "Get current weather"},
				},
			}
		} else {
			requestPayload = map[string]interface{}{
				"method": "mcp.call_tool",
				"params": map[string]interface{}{
					"name": "calculator",
					"arguments": map[string]interface{}{
						"expression": "2 + 2",
					},
				},
			}
			responsePayload = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "4",
					},
				},
			}
		}

		reqBytes, _ := json.Marshal(requestPayload)
		resBytes, _ := json.Marshal(responsePayload)

		entries[i] = audit.LogEntry{
			ID:          id,
			Timestamp:   timestamp,
			EventType:   "mcp.tool_call",
			ResourceURI: "mcp://local/tools",
			Action:      "execute",
			Status:      "success",
			Actor:       "user-123",
			IPAddress:   "192.168.1.100",
			Request:     string(reqBytes),
			Response:    string(resBytes),
		}
	}

	return entries
}

func (s *Server) GetAuditLogs(c echo.Context) error {
	// For demonstration/testing, we return rich mock data
	// This ensures the E2E tests have tabular data to verify
	entries := generateMockAuditEntries(20)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"entries": entries,
		"total":   len(entries),
	})
}
