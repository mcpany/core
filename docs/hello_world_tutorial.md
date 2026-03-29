# MCP Any - Hello World Tutorial

Welcome to the MCP Any Hello World tutorial! This guide illustrates the dramatically simplified setup experience when using `bazelisk run`.

## The "Day One" Experience

In the past, starting the server locally using Bazel introduced friction: because `bazel run` executes within a temporary runfiles directory, relative paths to configuration files provided on the CLI (like `examples/hello_world.yaml`) would result in a "Configuration file not found" error unless you specified absolute paths like `$(pwd)/examples/hello_world.yaml`.

We've completely eliminated this hurdle.

## Running the Server

MCP Any now intelligently detects when it is being launched via Bazel (by inspecting the `BUILD_WORKING_DIRECTORY` environment variable) and seamlessly resolves your relative paths back to the directory where you invoked the command.

This means you can simply run:

```bash
bazelisk run //server/cmd/mcpany -- run --config-path examples/hello_world.yaml
```

The server will locate `examples/hello_world.yaml`, parse your capabilities, and boot up the Universal Adapter immediately.

## What's Next?

With the friction removed, you can connect your client (such as Claude Desktop or any other MCP-compatible agent) to your new locally running service on `:50050` and easily experiment with modifying the `hello_world.yaml` tool definitions.
