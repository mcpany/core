# Service Doctor

The **Service Doctor** is a powerful interactive diagnostic tool designed to help you troubleshoot connectivity and health issues with your upstream services.

![Service Doctor](../screenshots/service-doctor.png)

## Overview

When an upstream service (like a GitHub integration or a PostgreSQL database) fails, it can be difficult to determine if the issue is due to network connectivity, authentication failures, or configuration errors. The Service Doctor runs a battery of real-time checks to pinpoint the root cause.

## Features

-   **Interactive Selection**: Choose any configured upstream service from a dropdown list.
-   **Real-time Diagnostics**: Runs immediate checks against the backend service.
    -   **TCP/HTTP Reachability**: Verifies if the host and port are accessible.
    -   **Authentication Check**: Validates API keys or tokens (if applicable).
    -   **Latency Measurement**: Reports the round-trip time for the health check.
-   **Detailed Reporting**: Provides clear "Healthy", "Degraded", or "Failed" status with actionable error messages.

## How to Use

1.  Navigate to **System Diagnostics** in the sidebar.
2.  Locate the **Service Doctor** panel.
3.  Select the service you want to diagnose from the dropdown menu.
4.  Click **Diagnose**.
5.  View the results card for status and details.
