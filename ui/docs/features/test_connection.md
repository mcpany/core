# Test Connection & Health Diagnostics

MCP Any provides a built-in "Test Connection" and Health Diagnostics feature for upstream services. This allows you to verify that your configuration (addresses, credentials, file paths, commands) is correct and the service is reachable before enabling it or when troubleshooting issues.

## How to use

1.  Navigate to **Upstream Services** in the sidebar.
2.  Locate the service you want to test.
3.  Click the **Troubleshoot** button (Activity icon) in the Service Detail header, or select **Diagnose** from the actions menu in the service list.

The system will run a series of diagnostic steps, including:
1.  **Configuration Validation**: Checks for common configuration errors.
2.  **Browser Connectivity**: Attempts to connect to the service directly from your browser (for HTTP/WebSocket services).
3.  **Backend Health Check**: Asks the MCP Any server to verify reachability and health of the upstream service.

## Supported Checks

The Backend Health Check covers the following upstream types:

*   **HTTP / GraphQL**: Performs a reachability check (HEAD/GET request) to the configured address.
*   **Filesystem**: Verifies that the configured root paths exist and are accessible by the server.
*   **Command Line / Stdio**: Verifies that the command executable exists in the system PATH or at the absolute path provided, and that the working directory exists.
*   **MCP Remote**: Verifies reachability of the HTTP endpoint or existence of the Stdio command.

## Screenshot

![Test Connection](../screenshots/test_connection.png)
