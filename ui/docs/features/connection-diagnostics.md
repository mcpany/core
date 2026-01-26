# Connection Diagnostics

## Overview

The Connection Diagnostics feature provides a comprehensive tool for troubleshooting connectivity issues with upstream MCP services. It allows administrators to run real-time health checks on configured services, providing detailed feedback on latency, configuration validity, and specific connection errors.

## Features

-   **Real-time Validation**: Tests connection to HTTP, Command Line, and Remote MCP services on demand.
-   **Step-by-Step Analysis**: Breaks down the validation process into clear steps (e.g., "Configuration Syntax", "Connectivity Check").
-   **Detailed Error Reporting**: Displays specific error messages (e.g., DNS resolution failures, authentication errors) to help pinpoint the root cause.
-   **Latency Measurement**: Shows the round-trip time for the connection check.

## Usage

1.  Navigate to the **Upstream Services** page.
2.  Select a service to view its details.
3.  Click the **Run Diagnostics** button in the header.
4.  A modal will appear showing the progress and results of the diagnostic tests.

## Visuals

![Connection Diagnostics](../screenshots/connection-diagnostics.png)
