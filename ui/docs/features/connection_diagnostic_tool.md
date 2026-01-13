# Service Status & Diagnostics

## Overview

Monitor the health and connection status of your upstream services directly from the **Service Details** page.

## Key Features

-   **Status Indicator**: Real-time Badge (Active/Disabled) indicating if the service is enabled in the gateway.
-   **Last Error Visibility**: If the service encounters connection errors (e.g., during registration or health checks), the error message is displayed on the service card.
-   **Toggle Availability**: Quickly enable or disable a service to stop traffic routing without uninstalling it.
-   **Connection Diagnostics**: A "Troubleshoot" wizard to actively probe the connection and validate configuration.

## Usage

1.  Navigate to **Services** (`/upstream-services`).
2.  Click on any service card to view its **Details**.
3.  Observe the **Status Badge** next to the title.
4.  Check for any **Error Alerts** in the overview section.

### Diagnostic Tool

If a service is unhealthy, click the **Troubleshoot** button to open the Connection Diagnostics dialog. This tool performs a multi-step check:
1.  **Client-Side Configuration**: Validates URL format and required fields.
2.  **Backend Connectivity**: Pings the backend to verify reachability and handshake status.
3.  **Logs**: Displays real-time diagnostic logs to help identify the root cause (DNS, Firewall, Auth).
