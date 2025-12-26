# MCP Any Manager Features

This document outlines the features implemented in the MCP Any Manager interface.

## Dashboard & Observability
**Status**: Implemented

The dashboard provides a real-time overview of the system's health and performance.

*   **Real-time Metrics**: Displays key performance indicators for Services, Tools, Resources, and Prompts.
*   **System Health**: Shows live health status (Healthy, Degraded, Unhealthy) for all connected upstream services.
*   **Aesthetics**: Designed with a glassmorphism effect and clean layout for a premium feel.

![Dashboard](.audit/ui/2025-05-23/dashboard.png)

## Core Management

### Services
**Status**: Implemented

Manage connected upstream services (HTTP, gRPC, Command Line, MCP Proxy).
*   **List View**: View all services with their type, version, and status.
*   **Toggle**: Enable/Disable services with a single click.
*   **Edit**: Configure service details.

![Services](.audit/ui/2025-05-23/services.png)

### Tools
**Status**: Implemented

View and manage tools exposed by connected services.
*   **Discovery**: Automatically lists tools from registered services.
*   **Control**: Enable or disable specific tools.

![Tools](.audit/ui/2025-05-23/tools.png)

### Resources
**Status**: Implemented

Manage access to resources (files, databases, etc.).
*   **Inventory**: View all available resources and their MIME types.
*   **Access Control**: Toggle access to specific resources.

![Resources](.audit/ui/2025-05-23/resources.png)

### Prompts
**Status**: Implemented

Manage prompt templates.
*   **Library**: View available prompts.
*   **Availability**: Enable or disable prompts for LLM use.

![Prompts](.audit/ui/2025-05-23/prompts.png)

## Advanced Features

### Middleware Pipeline
**Status**: Implemented

Visual management of the request processing pipeline.
*   **Drag & Drop**: Reorder middleware components (Auth, Rate Limit, Logging, etc.).
*   **Visualization**: See the flow of requests from ingress to service.

![Middleware](.audit/ui/2025-05-23/middleware.png)

### Webhooks
**Status**: Implemented

Configure outbound webhooks for system events.
*   **Configuration**: Add new webhook endpoints.
*   **Event Filtering**: Subscribe to specific events (e.g., service.registered, error.occurred).
*   **Testing**: Trigger test events to verify delivery.

![Webhooks](.audit/ui/2025-05-23/webhooks.png)
