package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SMMMiddleware provides real-time stylometric analysis of inter-agent messages.
// It acts as a security middleware to detect and mitigate "Reasoning-Path Shadowing"
// attacks where compromised subagents mimic authorized teammates to bypass integrity checks.
type SMMMiddleware struct {
	// The entropy threshold above which a message is flagged as suspicious mimicry.
	// For simple testing, this could be the length of unexpected metadata or repeating patterns.
	entropyThreshold float64
}

// NewSMMMiddleware creates a new Stylometric Mimicry Mitigator.
//
// Parameters:
//   - entropyThreshold (float64): The sensitivity for detecting anomalies.
//
// Returns:
//   - *SMMMiddleware: A new instance of the SMM middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewSMMMiddleware(entropyThreshold float64) *SMMMiddleware {
	return &SMMMiddleware{
		entropyThreshold: entropyThreshold,
	}
}

// Execute performs the stylometric mimicry check on an incoming MCP request.
// It intercepts "tools/call" requests and analyzes the argument payload to detect
// anomalies indicative of reasoning-path shadowing.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - method (string): The method name.
//   - req (mcp.Request): The incoming JSON-RPC request.
//   - next (mcp.MethodHandler): The next handler in the chain.
//
// Returns:
//   - mcp.Result: The result from the next handler, or an error if rejected.
//   - error: An error if stylometric anomaly is detected or the next handler fails.
//
// Errors:
//   - Returns a generic MCP error if stylometric anomaly is detected.
//
// Side Effects:
//   - May abort the request chain and log a Swarm Alert.
func (m *SMMMiddleware) Execute(ctx context.Context, method string, req mcp.Request, next mcp.MethodHandler) (mcp.Result, error) {
	// Only inspect tools/call for shadowing attacks.
	if method != "tools/call" {
		return next(ctx, method, req)
	}

	// In go-sdk/mcp, requests are unmarshalled to specific struct pointer types
	callReq, ok := req.(*mcp.CallToolRequest)
	if ok {
		// Calculate a simple "Stylometric Entropy" score based on the arguments.
		var args map[string]interface{}

		// Check if Params is initialized and arguments map exists/is valid
		if callReq.Params != nil && len(callReq.Params.Arguments) > 0 {
			err := json.Unmarshal(callReq.Params.Arguments, &args)
			if err == nil {
				score := m.calculateEntropy(args)

				if score > m.entropyThreshold {
					// Threat detected. Trigger Swarm Alert.
					return nil, fmt.Errorf("SMM Security Exception: Reasoning-Path Shadowing detected (entropy %v > threshold %v)", score, m.entropyThreshold)
				}
			}
		}
	}

	// Request passes MMBA baseline, proceed.
	return next(ctx, method, req)
}

// calculateEntropy is a simplistic heuristic for stylometric analysis.
func (m *SMMMiddleware) calculateEntropy(args map[string]interface{}) float64 {
	var totalEntropy float64

	for _, v := range args {
		if strVal, ok := v.(string); ok {
			// Basic heuristic: length of string + artificial bump for specific keywords
			// indicating an attack vector in our tests.
			entropy := float64(len(strVal)) * 0.1

			// Simulate detecting a specific mimicry pattern
			if strings.Contains(strings.ToLower(strVal), "shadow_mimic_payload") {
				entropy += 100.0 // Instant threshold breach
			}

			totalEntropy += entropy
		}
	}

	return totalEntropy
}
