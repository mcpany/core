# Debugging

MCP Any provides a debug mode that can be used to inspect the full request and response of each tool call. This is useful for troubleshooting your configurations and understanding the flow of data through the server.

## Enabling Debug Mode

You can enable debug mode by passing the `--debug` flag when starting the server:

```bash
./build/bin/server run --debug
```

Alternatively, you can set the `MCPANY_DEBUG` environment variable to `true`:

```bash
export MCPANY_DEBUG=true
./build/bin/server run
```

When debug mode is enabled, the server will log the full JSON-RPC request and response for each tool call. The logs will be printed to standard output or to the file specified by the `--logfile` flag.

## Example Log Output

Here is an example of the log output when debug mode is enabled (using structured logging):

```text
level=DEBUG msg="MCP Request" method=tools/list request="{\"jsonrpc\":\"2.0\",\"method\":\"tools/list\",\"id\":1}"
level=DEBUG msg="MCP Response" method=tools/list response="{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"tools\":[{\"name\":\"my-tool\"}]}}"
```

The log output includes the following information:

- **method**: The name of the MCP method that was called.
- **request**: The full JSON-RPC request.
- **response**: The full JSON-RPC response.

By inspecting the log output, you can identify any issues with your configuration and ensure that the server is behaving as expected.

## System Health & Diagnostics (Doctor)

For troubleshooting system startup, configuration issues, and upstream connectivity, MCP Any includes a built-in `doctor` command.

### Usage

```bash
# If using the server binary
./build/bin/server doctor

# If installed as 'mcpany'
mcpany doctor

# If using the mcpctl CLI
mcpctl doctor
```

### What it Checks

The doctor command performs a comprehensive health assessment in three stages:

1.  **Configuration Validation**:
    *   Loads all configuration files.
    *   Validates syntax and schema compliance.
    *   Checks for missing environment variables or invalid values.

2.  **Server Connectivity**:
    *   Attempts to connect to the running MCP Any server (default port 50050).
    *   Verifies the `/health` endpoint is responsive.

3.  **Deep Health Analysis**:
    *   Queries the internal `/doctor` endpoint.
    *   Reports the status of all registered upstream services.
    *   Highlights degraded or unhealthy services with specific error messages.

### Example Output

```text
Running Doctor checks...
========================
[ ] Checking Configuration... OK
[ ] Checking Server Connectivity (http://localhost:50050)... OK
[ ] Checking System Health... OK
    - my-postgres-service: OK
    - my-openai-service: OK
    - legacy-api: FAIL (connection refused) - Check if the upstream service is running.
All checks passed!
```
