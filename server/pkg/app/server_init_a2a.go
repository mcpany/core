package app

import (
	"github.com/mcpany/core/server/api"
)

// setupA2ARoutes attaches the A2A Messaging Hub to the application's HTTP multiplexer.
func (a *Application) setupA2ARoutes() {
	if a.standardMiddlewares != nil && a.standardMiddlewares.A2ABridge != nil {
		hub := api.NewA2AMessagingHub(a.standardMiddlewares.A2ABridge)
		hub.RegisterRoutes(a.Mux)
	}
}
