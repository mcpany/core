# UI Overhaul & Feature Implementation

This document details the UI overhaul and new features implemented for MCP Any.

## Dashboard & Observability

The dashboard provides real-time metrics and system health status.

### Features

- **Metrics Overview**: Displays total requests, active services, connected tools, and active users.
- **System Health**: Real-time status list of critical services with latency and uptime.
- **Request Volume**: A visual area chart showing request trends over time.

![Dashboard](screenshots/dashboard.png)

## Services Management

Manage upstream services (HTTP, gRPC, CMD, MCP).

### Features

- **List View**: View all registered services with their status, type, and version.
- **Toggle**: One-click enable/disable for services.
- **Create/Edit**: Form (Sheet) to register or update service configurations.

![Services](screenshots/services.png)

## Tools, Resources, and Prompts

Explore capabilities exposed by upstream services.

### Tools

List of available tools that can be invoked by the agent.
![Tools](screenshots/tools.png)

### Resources

Managed resources available to the system.
![Resources](screenshots/resources.png)

### Prompts

System prompts pre-configured for agents.
![Prompts](screenshots/prompts.png)

## Advanced Features

### Middleware Pipeline

Visual representation of the middleware execution order.
![Middleware](screenshots/middleware.png)

### Webhooks

Configure and test webhooks for system events.
![Webhooks](screenshots/webhooks.png)
