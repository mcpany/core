# Connection Diagnostics

MCP Any provides a powerful **Active Connection Diagnostics** tool to help you troubleshoot upstream service connectivity issues.

## Feature Overview

When a service is configured but fails to connect (e.g., due to network issues, authentication failures, or misconfiguration), it can be difficult to pinpoint the cause. The Diagnostics tool provides a step-by-step verification process.

### Capabilities

1.  **Client-Side Validation**: Checks for common configuration errors (e.g., malformed URLs).
2.  **Browser Connectivity Check**: Attempts to connect to the service directly from your browser. This helps identify issues where the server is reachable from your machine but not the backend, or vice versa (e.g., CORS, Firewall).
3.  **Active Backend Verification**: Triggers an immediate, real-time health check from the MCP Any backend server to the upstream service. This bypasses cached status and gives you the exact error returned by the connection attempt (e.g., "Connection Refused", "TLS Handshake Failed").

## How to Use

1.  Navigate to the **Services** page.
2.  Click the **Alert Icon** (triangle) next to a failing service, or click the **Actions** menu (three dots) and select **Diagnose**.
3.  The Diagnostics dialog will open.
4.  Click **Start Diagnostics**.
5.  Review the logs and status of each step.
6.  If an error is found, a suggestion will be displayed to help you fix it.

![Connection Diagnostics](../screenshots/connection-diagnostics.png)
