# System Diagnostics

The **System Diagnostics** page provides a centralized dashboard for monitoring the health and connectivity of the MCP Any server.

## Overview

Access the Diagnostics page via the **Diagnostics** link in the sidebar (Platform section).

![Diagnostics Page](../screenshots/diagnostics.png)

## Features

-   **System Status**: A high-level indicator (Healthy, Degraded, Unhealthy) based on aggregated checks.
-   **Core Checks**: Monitors internal server components (Version, Uptime, Memory).
-   **Environment**: Verifies the runtime environment (Docker, Node.js, Go).
-   **Network Connectivity**: Checks ability to reach external networks and upstream services.
-   **Service Health**: Real-time latency and status checks for all configured upstream services.

## Usage

1.  Navigate to **Diagnostics**.
2.  Click **Run Check** to refresh the diagnostic report.
3.  Review any degraded or failed checks.
4.  Detailed logs and diffs are provided for failed configuration checks.
