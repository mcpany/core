# Service Health Check (Test Connection)

MCP Any provides a built-in "Test Connection" feature for upstream services. This allows administrators to quickly verify if an upstream service (HTTP, gRPC, MCP, etc.) is reachable and healthy without leaving the dashboard.

## Overview

The health check feature is accessible directly from the **Services** page. It triggers a real-time connectivity check from the server to the upstream provider.

## Usage

1.  Navigate to the **Services** page.
2.  Locate the service you want to test.
3.  Click the **Actions** menu (three dots) on the right side of the service row.
4.  Select **Test Connection**.

## Behavior

-   **Success**: A green toast notification confirms the connection was successful.
-   **Failure**: A red toast notification appears with details about the error (e.g., "Connection refused", "Timeout").

## Implementation Details

-   **HTTP/GraphQL**: Performs a HEAD or GET request to the service URL.
-   **gRPC**: Checks if the connection pool can establish a connection.
-   **MCP (Stdio/SSE)**: Verifies the client session is active and responsive.

![Services Page](../screenshots/services.png)
