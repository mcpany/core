# MCP Any UI Features

This document describes the key features of the MCP Any management interface.

## Dashboard

The Dashboard provides a real-time overview of the system status, including:

- Total Requests & Trends
- Active Services count
- Connected Tools & Resources
- Latency and Error Rate metrics

![Dashboard Screenshot](screenshots/dashboard.png)

## Services Management

Manage all connected Upstream Services.

- **List View**: See all services, their type (HTTP, gRPC, MCP, CMD), version, and status.
- **Toggle**: Enable or disable services instantly.
- **Edit**: Configure connection details and endpoints.

![Services Screenshot](screenshots/services.png)

## Tools

View and manage all tools exposed by the upstream services.

- Search and filter tools.
- Enable/Disable individual tools.

![Tools Screenshot](screenshots/tools.png)

## Middleware

Configure the request processing pipeline.

- Drag-and-drop interface to reorder middleware.
- Enable/Disable middleware components (Auth, Rate Limiting, Logging, etc.).
- Visualize the flow of requests.

![Middleware Screenshot](screenshots/middleware.png)

## Webhooks

Configure outbound webhooks for system events.

- Register new webhook endpoints.
- Select events to subscribe to.
- Test and monitor webhook delivery status.

![Webhooks Screenshot](screenshots/webhooks.png)
