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

When debug mode is enabled, the server will log detailed information about each tool call, including source code locations. It logs the start and completion (or failure) of requests, along with duration.

## Example Log Output

Here is an example of the log output when debug mode is enabled (using structured logging):

```text
level=INFO msg="Request completed" method=tools/list duration=1.2ms
level=ERROR msg="Request failed" method=tools/call duration=45.2ms error="tool execution failed"
```

The log output includes the following information:

- **method**: The name of the MCP method that was called.
- **duration**: The time taken to process the request.
- **error**: The error message if the request failed.

By inspecting the log output, you can identify any issues with your configuration and ensure that the server is behaving as expected.
