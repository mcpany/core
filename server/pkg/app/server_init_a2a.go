package app

import (
	"net/http"

	"github.com/mcpany/core/server/api"
)

// setupA2ARoutes attaches the Agent-to-Agent (A2A) Messaging Hub to the
// application's HTTP multiplexer if the bridge middleware is available.
//
// Parameters:
//   - mux (*http.ServeMux): The application's HTTP multiplexer.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Registers /v1/a2a routes on the application's mux if configured.
func (a *Application) setupA2ARoutes(mux *http.ServeMux) {
	if a.standardMiddlewares != nil && a.standardMiddlewares.A2ABridge != nil {
		hub := api.NewA2AMessagingHub(a.standardMiddlewares.A2ABridge)
		hub.RegisterRoutes(mux)
	}
}
