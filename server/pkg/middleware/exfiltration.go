package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

// ExfiltrationTransportGateway is a middleware that prevents API key exfiltration
// by enforcing an allow-list of upstream domains.
type ExfiltrationTransportGateway struct {
	allowList map[string]bool
}

// NewExfiltrationTransportGateway creates a new ExfiltrationTransportGateway.
func NewExfiltrationTransportGateway(allowedDomains []string) *ExfiltrationTransportGateway {
	allowMap := make(map[string]bool)
	for _, d := range allowedDomains {
		allowMap[strings.ToLower(d)] = true
	}
	return &ExfiltrationTransportGateway{
		allowList: allowMap,
	}
}

// Handle implements the HTTP middleware interface.
func (m *ExfiltrationTransportGateway) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger().With("component", "exfiltration_transport")

		targetHost := r.Host

		// If using X-Forwarded-Host or similar proxy headers
		if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
			targetHost = xfh
		}

		// Also check the URL Host if present
		if r.URL.Host != "" {
			targetHost = r.URL.Host
		}

		// Clean the host (remove port if present)
		cleanHost := targetHost
		if host, _, err := net.SplitHostPort(targetHost); err == nil {
			cleanHost = host
		}
		cleanHost = strings.ToLower(cleanHost)

		// Is the host allowed?
		allowed := false
		if m.allowList[cleanHost] {
			allowed = true
		} else {
			// Check wildcards
			for allowedDomain := range m.allowList {
				if strings.HasPrefix(allowedDomain, "*.") {
					suffix := allowedDomain[1:] // e.g., ".anthropic.com"
					if strings.HasSuffix(cleanHost, suffix) {
						allowed = true
						break
					}
				}
			}
		}

		// Allow local traffic for the gateway itself
		if cleanHost == "localhost" || cleanHost == "127.0.0.1" || cleanHost == "::1" {
			allowed = true
		}

		if !allowed {
			logger.WarnContext(r.Context(), "Blocked exfiltration attempt", "host", cleanHost)
			http.Error(w, fmt.Sprintf("Exfiltration Transport Gateway: Host %s is not in the allow-list", cleanHost), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
