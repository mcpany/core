# MCP Any UI - Feature Documentation

## Overview
The MCP Any UI serves as the central management console for the MCP Any server. It allows administrators to monitor server health, manage upstream services, explore available tools/resources/prompts, and configure advanced settings like webhooks and middleware.

## Dashboard
The dashboard provides a high-level overview of the server's status.
- **Metrics:** Real-time visualization of Total Requests, Active Services, Latency, and Users.
- **Service Health:** Live status (Healthy, Degraded, Unhealthy) of all connected upstream services.

![Dashboard](.audit/ui/2025-12-24/dashboard.png)

## Services Management
The Services page allows full control over upstream integrations.
- **List:** View all registered services, their types (HTTP, gRPC, etc.), and versions.
- **Toggle:** Enable or disable services instantly.
- **Edit:** Configure service details (Name, Version, Connection settings).

![Services](.audit/ui/2025-12-24/services.png)

## Tools, Resources, Prompts
Dedicated pages allow exploration of the capabilities exposed by the server.
- **Tools:** List of all functions available to LLMs.
- **Resources:** Static or dynamic context data.
- **Prompts:** Pre-defined interaction templates.

![Tools](.audit/ui/2025-12-24/tools.png)

## Advanced Configuration
### Webhooks
Manage event subscriptions to trigger external actions based on server events.

![Webhooks](.audit/ui/2025-12-24/webhooks.png)

### Middleware
Visualize and manage the request processing pipeline. Enable/Disable middleware components and view their priority.

![Middleware](.audit/ui/2025-12-24/middleware.png)
