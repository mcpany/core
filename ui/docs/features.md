
# MCP Any UI Features

This document provides an overview of the features available in the MCP Any management console.

## 1. Dashboard & Observability

The dashboard provides a real-time overview of the system's health and performance.

*   **Key Metrics:** Active Services, Requests/sec, Average Latency, and Active Resources.
*   **Service Health:** A live table showing the status (Active/Inactive/Warning) of all connected upstream services.
*   **Request Volume:** A chart visualizing traffic over the last 24 hours.

![Dashboard]( ../../.audit/ui/2026-01-07/Dashboard.png)

## 2. Service Management

Manage your upstream MCP services (HTTP, gRPC, Command Line, MCP Proxy).

*   **List View:** View all services with their status, type, and priority.
*   **Toggle:** Enable or disable services with a single click.
*   **CRUD:** Add new services or edit existing configurations using the slide-out sheet.

![Services]( ../../.audit/ui/2026-01-07/Services.png)

## 3. Tool Explorer

Browse and test tools exposed by your connected MCP servers.

*   **Discovery:** Automatically lists tools available across all active services.
*   **Testing:** built-in JSON editor to execute tools with custom arguments and view the output.

![Tools]( ../../.audit/ui/2026-01-07/Tools.png)

## 4. Resource Management

View data resources available to LLMs.

*   **List View:** See resource URIs and MIME types.
*   **Quick Copy:** Easily copy resource URIs for use in prompts or testing.

![Resources]( ../../.audit/ui/2026-01-07/Resources.png)

## 5. Prompt Templates

Manage and execute reusable prompt templates.

*   **Template Library:** Browse available prompts with descriptions and argument requirements.
*   **Execution:** Run prompts directly from the UI by filling out a generated form based on the template arguments.

![Prompts]( ../../.audit/ui/2026-01-07/Prompts.png)

## 6. Profiles

Manage execution environments and access controls.

*   **Environments:** Configure profiles for Development, Production, or Debugging.
*   **Access Control:** Assign roles and permissions to specific profiles.

![Profiles]( ../../.audit/ui/2026-01-07/Profiles.png)

## 7. Middleware Pipeline

Visually manage the request processing pipeline.

*   **Drag-and-Drop:** Reorder middleware to change the execution priority.
*   **Toggle:** Enable/Disable specific middleware components (e.g., Logging, Rate Limiting) without restarting the server.

![Middleware]( ../../.audit/ui/2026-01-07/Middleware.png)

## 8. Webhooks

Configure external notifications for server events.

*   **Management:** Add, edit, and remove webhook endpoints.
*   **Testing:** Send test payloads to verify webhook connectivity directly from the UI.

![Webhooks]( ../../.audit/ui/2026-01-07/Webhooks.png)
