# Test Connection & Health Diagnostics

MCP Any provides a built-in "Test Connection" feature for upstream services. This allows you to verify that your configuration (addresses, credentials, file paths, commands) is correct and the service is reachable before enabling it or when troubleshooting issues.

## How to use

1.  Navigate to **Upstream Services** in the sidebar.
2.  Click on the service you want to test.
3.  In the service detail header, click the **Test Connection** button (play icon).

The system will attempt to connect to the service or verify the configuration resources and display the result.

## Supported Checks

*   **HTTP / GraphQL**: Performs a reachability check (HEAD/GET request) to the configured address.
*   **Filesystem**: Verifies that the configured root paths exist and are accessible by the server.
*   **Command Line / Stdio**: Verifies that the command executable exists in the system PATH or at the absolute path provided, and that the working directory exists.
*   **MCP Remote**: Verifies reachability of the HTTP endpoint or existence of the Stdio command.

## Screenshot

![Test Connection](../screenshots/test_connection.png)
