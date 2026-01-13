# Service Connection Diagnostic Tool

## Overview

A new diagnostic tool has been added to the Service Details page. This tool helps users troubleshoot connection issues with their upstream MCP services by validating the configuration and reporting the current connection status from the backend.

## Key Features

-   **Client-Side Configuration Check**: Validates that the service URL or command is correctly formatted before attempting to connect.
-   **Backend Status Check**: Queries the MCP Any backend for the real-time health status of the service.
-   **Error Hints**: Provides actionable hints for common errors like connection refusal, timeouts, or handshake failures.

## Usage

1.  Navigate to any **Service Detail** page.
2.  Click the **"Troubleshoot"** button in the header.
3.  In the dialog, click **"Start Diagnostics"**.
4.  View the real-time logs and status of the checks.

## Screenshot

(Screenshot placeholder - feature verified via tests)
