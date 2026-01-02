# MCP Any UI Features

This document outlines the key features of the MCP Any management console.

## Dashboard & Observability
**Screenshot:** `.audit/ui/{YYYY-MM-DD}/dashboard.png`

The dashboard provides real-time visibility into the health and performance of your MCP ecosystem.
*   **Key Metrics:** View Total Requests, Active Services, Connected Tools, and more at a glance.
*   **Health Status:** Immediate visual feedback on system status.
*   **Latency & Errors:** Monitor performance trends with sparklines and trend indicators.

## Services Management
**Screenshot:** `.audit/ui/{YYYY-MM-DD}/services.png`

Manage your upstream services (HTTP, gRPC, MCP, CMD, etc.).
*   **List View:** See all registered services, their type, version, and status.
*   **Toggle:** Enable or disable services with a single click.
*   **Configuration:** Edit service connection details (Endpoint, Command, etc.) via a slide-out sheet.

## Tools Management
**Screenshot:** `.audit/ui/{YYYY-MM-DD}/tools.png`

Explore and manage tools exposed by your upstream services.
*   **Discovery:** Automatically lists tools from connected services.
*   **Inspection:** View tool schemas and details.
*   **Control:** Enable/Disable specific tools.

## Resources & Prompts
**Screenshots:**
*   `.audit/ui/{YYYY-MM-DD}/resources.png`
*   `.audit/ui/{YYYY-MM-DD}/prompts.png`

View and manage resources and prompts provided by your MCP servers.

## Middleware Pipeline
**Screenshot:** `.audit/ui/{YYYY-MM-DD}/middleware.png`

Visually manage the request processing pipeline.
*   **Drag & Drop:** Reorder middleware (Auth, Rate Limiting, Logging) using a drag-and-drop interface.
*   **Toggle:** Enable/Disable specific middleware components.
*   **Visualization:** See the request flow from ingress to service.

## Webhooks
**Screenshot:** `.audit/ui/{YYYY-MM-DD}/webhooks.png`

Configure and test webhooks for event-driven integrations.
