# Upstream Service Logs

View the standard error (stderr) output from your upstream MCP services directly in the dashboard.

## Overview

When connecting to upstream MCP services, especially those running via standard I/O (stdio) or Docker, initialization errors can be hard to debug. The **Service Logs** tab captures the stderr output from the connection attempt, allowing you to quickly diagnose issues like missing dependencies, configuration errors, or crashes.

## How to Access

1.  Navigate to **Services** in the sidebar.
2.  Click on any service to view its details.
3.  Select the **Logs** tab.

## Features

-   **Connection Output**: See exactly what the upstream service printed to stderr during startup.
-   **Debug Info**: Helpful for diagnosing `python: command not found`, `ModuleNotFoundError`, or JSON-RPC handshake failures.

![Service Logs](../screenshots/service_config.png)
*(Note: Select the "Logs" tab in the service detail view)*
