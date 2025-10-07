# MCPXY Examples

This directory contains examples of how to use MCPXY with different upstream service types. Each example includes:

- An upstream service implementation.
- A configuration file for MCPXY.
- A shell script to start an MCPXY server configured for the example.
- A README file with instructions on how to run the example and interact with it using AI tools like Gemini CLI.

## 1. Build the MCPXY binary

Before running any of the examples, you need to build the `mcpxy` binary. From the root of the `core` project, run:

```bash
make build
```

This will create the `mcpxy` binary in the `build/bin` directory.

## 2. Running the Examples

Each example has a `start.sh` script that starts the MCPXY server with the correct configuration. The upstream service needs to be started separately. Please refer to the `README.md` file in each example's directory for detailed instructions.

## 3. Interacting with the Examples using AI Tools

The examples are designed to be used with AI tools that can consume MCPXY extensions. Here's a general guide on how to use them with different tools.

### Gemini CLI

1.  **Configure the Extension:** Open your Gemini CLI configuration file (e.g., `~/.config/gemini/config.yaml`) and add an extension for the example you want to use. For example, for the HTTP time server example:

    ```yaml
    extensions:
      mcpxy-http-time:
        http:
          address: http://localhost:8080
    ```

2.  **List Available Tools:** Use the `gemini list tools` command to see the tools exposed by the MCPXY server.

3.  **Call a Tool:** Use the `gemini call tool` command to interact with the service.

### Claude Desktop

(Instructions for Claude Desktop would go here. I am not familiar with Claude Desktop, so I will leave this as a placeholder.)

### VS Code

(Instructions for VS Code extensions would go here. I am not familiar with VS Code extensions that consume MCPXY, so I will leave this as a placeholder.)

### Other Tools

The general principle is to find the tool's configuration file and add an HTTP extension pointing to the MCPXY server's address (usually `http://localhost:8080` in these examples).
