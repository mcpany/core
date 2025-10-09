# 🔌 Integrating with AI Assistants

This guide explains how to connect `mcpxy` to various AI assistant clients, enabling them to leverage the tools exposed by your `mcpxy` server. By integrating `mcpxy`, you can give your AI assistant access to gRPC services, REST APIs, local command-line tools, and more.

## Prerequisites

Before you can integrate `mcpxy` with an AI assistant, you need to have the `mcpxy` server binary built and available on your system.

1.  **Clone the `mcpxy` repository:**
    ```bash
    git clone https://github.com/mcpxy/core.git
    cd core
    ```

2.  **Build the application:**
    This command compiles the server and places the executable binary in the `./bin` directory.
    ```bash
    make build
    ```
    The server binary will be located at `./bin/server`. You will need to use the **absolute path** to this binary when configuring your AI assistant. For example, if you cloned the repository to `/home/user/mcpxy-core`, the path would be `/home/user/mcpxy-core/bin/server`.

## General Configuration

Most MCP clients require you to specify a command to start the MCP server. For `mcpxy`, this command is the absolute path to the `server` binary you built. You can also pass any of the `mcpxy` command-line arguments.

Here is a generic JSON configuration that can be adapted for many clients. This example shows how to point to a configuration file.

```json
{
  "mcpServers": {
    "mcpxy": {
      "command": "/path/to/your/mcpxy-core/bin/server",
      "args": [
        "--config-paths",
        "/path/to/your/mcpxy-config.yaml"
      ]
    }
  }
}
```

**Note:**
*   Replace `/path/to/your/mcpxy-core/bin/server` with the actual absolute path to the `server` binary on your filesystem.
*   If you want to register tools dynamically without a static config file, you can omit the `args` array.

## AI Client Setup Examples

Below are specific instructions for popular AI assistants.

### Gemini CLI

You can register `mcpxy` as an extension to the Gemini CLI, making all its tools available in your chat sessions.

**To add `mcpxy` for the current project:**
```bash
gemini mcp add mcpxy /path/to/your/mcpxy-core/bin/server
```

**To add `mcpxy` globally for your user:**
```bash
gemini mcp add -s user mcpxy /path/to/your/mcpxy-core/bin/server
```

**To add `mcpxy` with command-line arguments (like a config file):**
Use the `--` separator to pass arguments to the server command.
```bash
gemini mcp add mcpxy -- /path/to/your/mcpxy-core/bin/server --config-paths /path/to/your/mcpxy-config.yaml
```

### Claude Code

Use the `claude` CLI to add the `mcpxy` server:
```bash
claude mcp add mcpxy /path/to/your/mcpxy-core/bin/server
```

### Copilot CLI

1.  Start the Copilot CLI in your terminal:
    ```bash
    copilot
    ```
2.  Run the command to add a new MCP server:
    ```
    /mcp add
    ```
3.  Configure the fields in the interactive prompt:
    *   **Server name:** `mcpxy`
    *   **Server Type:** `Local`
    *   **Command:** `/path/to/your/mcpxy-core/bin/server`
    *   **Arguments:** (Optional) `--config-paths`, `/path/to/your/mcpxy-config.yaml`
4. Press `CTRL+S` to save.

### VS Code / Copilot Chat

Follow the official guide for [adding an MCP server in VS Code](https://code.visualstudio.com/docs/copilot/chat/mcp-servers#_add-an-mcp-server) and use the JSON configuration from the "General Configuration" section above.

### Cursor

1.  Go to `Cursor Settings` -> `MCP` -> `New MCP Server`.
2.  Use the JSON configuration from the "General Configuration" section above.

### JetBrains AI Assistant & Junie

1.  Go to `Settings` | `Tools` | `AI Assistant` | `Model Context Protocol (MCP)`.
2.  Click `Add` and use the JSON configuration from the "General Configuration" section above.
3.  The same process applies for Junie under `Settings` | `Tools` | `Junie` | `MCP Settings`.

---

By following these instructions, you can connect `mcpxy` to your favorite AI coding assistant and extend its capabilities significantly.