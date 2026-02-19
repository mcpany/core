# Services Management

**Status:** Implemented

## Goal

Manage upstream services connected to the MCP Any platform. The Services Dashboard allows operators to view the health of connected services, enable or disable them, and configure connection details including authentications and environment variables.

## Actors

- **Admin**: Has full access to manage services.
- **Operator**: Can view status and toggle services.

## Usage Guide

### 1. View Service List

Navigate to the **Services** page via the sidebar. This view provides a high-level overview of all registered upstream services.

![Services List](screenshots/services_list.png)

Key columns:

- **Name**: Application identifier and link to details.
- **Status**: Toggle switch to enable/disable traffic, plus diagnostic status.
- **Type**: Protocol (HTTP, gRPC, CMD, MCP).
- **Activity**: Sparkline visualization of recent latency.
- **Address**: The endpoint or command.
- **Actions**: Menu with options like "View Logs", "Diagnose", "Edit", and "Delete".

### 2. Add New Service

To register a new upstream service:

1. Click the **"Register Service"** button.
2. A **Configuration Dialog** will open.
3. Select a service template (e.g., PostgreSQL, HTTP) or configure a custom type.
4. Follow the configuration wizard to register the service.

![Add Service Dialog](screenshots/services_add_dialog.png)

### 3. Configure Service

To edit an existing service:

1. Click on the service name in the list (or select "Edit" from the Actions menu).
2. A detailed **Configuration Dialog** will open.
3. Here you can update the endpoint, managing **Environment Variables**, and view specialized settings.

![Service Configuration](screenshots/service_config.png)

### 4. Toggle Service State

You can instantly stop routing traffic to a service by toggling the switch in the main list.

- **On**: Service is active and receiving traffic.
- **Off**: Service is disconnected; dependent tools will be unavailable.

## Technical Details

### Supported Service Types

- **HTTP (OpenAPI)**: Connects to REST/OpenAPI endpoints. Ideal for third-party SaaS (e.g., GitHub, Stripe).
- **gRPC**: Connects to high-performance internal microservices using Protobuf reflection.
- **MCP**: Connects to other Model Context Protocol servers.
- **CMD (Local)**: Executes local command-line tools (stdio). Perfect for scripts, Python environments, or CLI utilities.

### Special Configuration

- **Environment Variables**: Define key-value pairs injected into the process (for CMD) or sent as metadata. Supports `secrets.*` references.
- **Health Checks**: The system periodically pings the `health_check_endpoint` (default `/health`) to update the status.
