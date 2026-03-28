package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

// CMCSProviderMiddleware implements Cross-Mesh Command Sovereignty.
// It validates "Mesh Tokens" for inter-teammate mailbox validation in horizontal swarms.
type CMCSProviderMiddleware struct {
	// A map of role -> allowed commands (mock)
	rolePolicies map[string][]string
}

// NewCMCSProviderMiddleware creates a new CMCSProviderMiddleware.
func NewCMCSProviderMiddleware(policies map[string][]string) *CMCSProviderMiddleware {
	return &CMCSProviderMiddleware{
		rolePolicies: policies,
	}
}

// Handle implements the HTTP middleware interface.
func (m *CMCSProviderMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger().With("component", "cmcs_provider")

		// Mesh Token Header expected format: "role:command" for testing simplicity
		// In reality, this would be a cryptographically signed token.
		meshToken := r.Header.Get("X-Mesh-Token")
		if meshToken != "" {
			parts := strings.Split(meshToken, ":")
			if len(parts) == 2 {
				role := parts[0]
				command := parts[1]

				allowedCommands, ok := m.rolePolicies[role]
				if !ok {
					logger.WarnContext(r.Context(), "CMCS validation failed: unknown role", "role", role)
					http.Error(w, fmt.Sprintf("CMCS: Unauthorized role '%s'", role), http.StatusForbidden)
					return
				}

				authorized := false
				for _, c := range allowedCommands {
					if c == command || c == "*" {
						authorized = true
						break
					}
				}

				if !authorized {
					logger.WarnContext(r.Context(), "CMCS validation failed: unauthorized command", "role", role, "command", command)
					http.Error(w, fmt.Sprintf("CMCS: Role '%s' not authorized for command '%s'", role, command), http.StatusForbidden)
					return
				}
			} else {
				logger.WarnContext(r.Context(), "Invalid X-Mesh-Token format", "token", meshToken)
				http.Error(w, "CMCS: Invalid Mesh Token format", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
