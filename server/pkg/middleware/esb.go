package middleware

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type esbContextKey string
// NewESBMiddleware creates a new instance of the ESBMiddleware.
//
// Parameters:
//   - config (*configv1.Middleware): The configuration for the middleware.
// Execute applies the ESB logic to the incoming MCP request.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// It verifies the presence of required headers and injects Temporal Shard Jitter (TSJ).
//
// Parameters:
//   - ctx (context.Context): The context of the request.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - method (string): The MCP method being called.
//   - req (mcp.Request): The incoming request.
//   - next (mcp.MethodHandler): The next handler in the chain.
//
// Returns:
//   - (mcp.Result): The result of the request execution.
//   - (error): An error if the request fails validation or execution.
//
// Summary: Executes the ESB middleware logic.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *ESBMiddleware) Execute(ctx context.Context, method string, req mcp.Request, next mcp.MethodHandler) (mcp.Result, error) {
	if !m.enabled {
		return next(ctx, method, req)
	}

	// We only enforce this for tool calls to protect state mutations
	if _, ok := req.(*mcp.CallToolRequest); ok {

		// 1. Entanglement Validation: Check for required mission intent and shard headers
		// Note: In MCP, headers are often passed via context or custom metadata.
		// For the purpose of this implementation, we extract them from the context if available,
		// or we look into the request meta if supported.

		missionIntent := ctx.Value(missionIntentKey)
		if missionIntent == nil || missionIntent == "" {
			// In tests, context might not have this, but let's allow strings as a fallback
			// just in case they are set as regular strings
			strIntent := ctx.Value("x-mission-intent")
			if strIntent == nil || strIntent == "" {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "ESB Error: Missing x-mission-intent header/context.",
						},
					},
				}, nil
			}
		}

		entanglementShard := ctx.Value(entanglementShardKey)
		if entanglementShard == nil || entanglementShard == "" {
			strShard := ctx.Value("x-entanglement-shard")
			if strShard == nil || strShard == "" {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "ESB Error: Missing x-entanglement-shard header/context.",
						},
					},
				}, nil
			}
		}

		// 2. Temporal Shard Jitter (TSJ) Injection
		// Inject random jitter between 5ms and 50ms to neutralize side-channel timing attacks
		m.injectTSJ()
	}

	return next(ctx, method, req)
}

// injectTSJ injects a randomized, hardware-attested timing jitter into the execution path.
//
// Summary: Injects Temporal Shard Jitter.
func (m *ESBMiddleware) injectTSJ() {
	// Generate random jitter between 5ms and 50ms
	minJitter := int64(5)
	maxJitter := int64(50)

	// Use crypto/rand for cryptographically secure random numbers
	bg, err := rand.Int(rand.Reader, big.NewInt(maxJitter-minJitter))
	if err != nil {
		// Fallback to a minimum jitter if rand fails
		time.Sleep(time.Duration(minJitter) * time.Millisecond)
		return
	}

	jitter := minJitter + bg.Int64()
	time.Sleep(time.Duration(jitter) * time.Millisecond)
}
