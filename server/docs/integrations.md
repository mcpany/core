# 🔌 Integrating with AI Assistants

This guide explains how to connect `mcpany` to various AI assistant clients, enabling them to leverage the tools exposed by your `mcpany` server. By integrating `mcpany`, you can give your AI assistant access to gRPC services, REST APIs, local command-line tools, and more.

1. **As a Local Binary**: Run the `mcpany` server directly on your machine.
2. **With Docker Compose**: Run `mcpany` as a containerized service.
3. **With Helm**: Deploy `mcpany` to a Kubernetes cluster.

> [!TIP] > **Best Practice: One Server, Many Tools**
> You don't need to run a separate `mcpany` instance for every tool you want to use. Instead, point your initialized `mcpany` server to multiple configuration files or a directory of configs using `--config-paths`. This way, you only need to register **one** MCP server with your AI assistant to access **all** your tools.

---

## 1. Local Binary Integration

This is the most direct method for running `mcpany` on your local machine.

### Prerequisites

Before you begin, you need to have the `mcpany` server binary built and available on your system.

1. **Clone the `mcpany` repository:**

   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```

2. **Build the application:**
   This command compiles the server and places the executable binary in the `build/bin` directory.

   ```bash
   make build
   ```

   The server binary will be located at `./build/bin/server`. You will need to use the **absolute path** to this binary when configuring your AI assistant.

### AI Client Setup

Most AI clients require you to specify the command to start the `mcpany` server.

#### General JSON Configuration

This generic JSON configuration can be adapted for most clients that support it (e.g., VS Code, JetBrains, Cursor).

```json
{
  "mcpServers": {
    "mcpany": {
      "command": "./build/bin/server",
      "args": ["--config-paths", "/path/to/your/mcpany-config.yaml"]
    }
  }
}
```

- **Note**: The server binary is now located at `./build/bin/server`.

#### Gemini CLI

To register `mcpany` as an extension to the Gemini CLI:

```bash
# Add for the current project
gemini mcp add mcpany "./build/bin/server"

# Add with command-line arguments (like a config file)
# Note the use of '--' to separate the command from its arguments.
gemini mcp add mcpany -- "./build/bin/server" --config-paths "/path/to/your/mcpany-config.yaml"
```

#### Claude CLI

```bash
claude mcp add mcpany "./build/bin/server"
```

#### Copilot CLI

Use the interactive prompt (`/mcp add`) and provide the absolute path to the `server` binary and any optional arguments.

---

## 2. Docker Compose Integration

You can run `mcpany` and its upstream services in a containerized environment using Docker Compose. This is ideal for creating reproducible setups.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) are installed.

### Running with Docker Compose

The repository includes a `docker-compose.yml` file that starts `mcpany` and a sample upstream service.

1. **Start the services:**

   ```bash
   docker compose up --build
   ```

   The `mcpany` server will be available on `localhost:50050`.

2. **Configure your AI Assistant:**
   Since `mcpany` is running as a networked service (not a local process), you need to configure your AI assistant to connect to its TCP port.

   For clients like Gemini, you can add it as an HTTP extension:

   ```bash
   # The 'mcp' subcommand is for local processes. For networked services, use 'http'.
   gemini http add mcpany-docker http://localhost:50050
   ```

   Or, to run it in **stdio mode** via Docker (slower than HTTP mode, but more secure):

   ```bash
   gemini mcp add --transport stdio mcpany-docker docker -- run -i --rm -v /absolute/path/to/config.yaml:/etc/mcpany/config.yaml ghcr.io/mcpany/server:latest run --config-path /etc/mcpany/config.yaml --stdio
   ```

   For clients with a UI (like VS Code or JetBrains), look for an option to add an "HTTP" or "Remote" MCP server and point it to `http://localhost:50050`.

---

## 3. Helm Integration

For production or staging environments, you can deploy `mcpany` to a Kubernetes cluster using the provided Helm chart.

### Prerequisites

- A running Kubernetes cluster.
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured to connect to your cluster.
- [Helm](https://helm.sh/docs/intro/install/) installed.

### Deployment

1. **Install the Helm Chart:**
   Navigate to the `helm` directory in this repository and install the chart.

   ```bash
   cd helm/
   helm install mcpany . -f values.yaml
   ```

   This will deploy `mcpany` into your Kubernetes cluster.

2. **Expose the Service:**
   To access `mcpany` from your local machine, you can use `kubectl port-forward`.

   ```bash
   # Forward a local port (e.g., 50050) to the mcpany service in the cluster
   kubectl port-forward svc/mcpany 50050:50050
   ```

3. **Configure your AI Assistant:**
   With the port forward running, `mcpany` is now accessible at `localhost:50050`. You can configure your AI assistant the same way as in the Docker Compose example.

   ```bash
   gemini http add mcpany http://localhost:50050
   ```

---

## 4. Claude Desktop

To use `mcpany` with [Claude Desktop](https://modelcontextprotocol.io/quickstart/user), configure it to run the `mcpany` Docker container.

Add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mcpany": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "/absolute/path/to/your/config.yaml:/etc/mcpany/config.yaml",
        "ghcr.io/mcpany/server:latest",
        "run",
        "--config-path",
        "/etc/mcpany/config.yaml",
        "--stdio"
      ]
    }
  }
}
```

> **Note:** Replace `/absolute/path/to/your/config.yaml` with the actual path to your MCP Any configuration file.

This setup ensures Claude launches a fresh `mcpany` container whenever it needs to access your tools.

---

## 5. Claude Code CLI

To integrate `mcpany` with the [Claude Code CLI](https://www.npmjs.com/package/@anthropic-ai/claude-code):

```bash
# Register mcpany
claude mcp add mcpany --transport http http://localhost:50050
```

---

## 6. GitHub Copilot CLI

The GitHub Copilot CLI does not currently support a non-interactive `mcp add` command for remote servers. Instead, you should configure it using an `mcp-config.json` file.

1.  Create an `mcp-config.json` file in your `~/.copilot` directory (or a directory specified by `XDG_CONFIG_HOME`).

```json
{
  "mcpServers": {
    "mcpany": {
      "url": "http://localhost:50050",
      "type": "http"
    }
  }
}
```

2.  Run the Copilot CLI as usual:

```bash
gh copilot explain "how does this code work?"
```

---

## 7. VS Code (GitHub Copilot)

You can add `mcpany` to GitHub Copilot in VS Code using the command line:

```bash
code --add-mcp '{"name":"mcpany","command":"docker","args":["run","-i","--rm","-v","/absolute/path/to/your/config.yaml:/etc/mcpany/config.yaml","ghcr.io/mcpany/server:latest","run","--config-path","/etc/mcpany/config.yaml","--stdio"]}'
```

Alternatively, you can manually edit your VS Code `settings.json` or use the **"Add MCP Server"** command in the command palette.

---

## 8. Cursor

To integrate with [Cursor](https://cursor.sh/):

1.  Go to **Cursor Settings** > **MCP**.
2.  Click **Add new MCP server**.
3.  Fill in the details:
    - **Name**: `mcpany`
    - **Type**: `command`
    - **Command**: `docker run -i --rm -v /absolute/path/to/your/config.yaml:/etc/mcpany/config.yaml ghcr.io/mcpany/server:latest run --config-path /etc/mcpany/config.yaml --stdio`
4.  Click **Add**.

---

## 9. JetBrains AI Assistant

To integrate with JetBrains IDEs (IntelliJ IDEA, PyCharm, GoLand, etc.):

1.  Open **Settings** (or **Preferences** on macOS).
2.  Navigate to **Tools** > **AI Assistant** > **Model Context Protocol (MCP)**.
3.  Click the **+** (Add) button.
4.  Configure the server:
    - **Name**: `mcpany`
    - **Command**: `docker`
    - **Args**: `run -i --rm -v /absolute/path/to/your/config.yaml:/etc/mcpany/config.yaml ghcr.io/mcpany/server:latest run --config-path /etc/mcpany/config.yaml --stdio`
5.  Click **OK** or **Apply**.

---

## 10. Cline

To use `mcpany` with [Cline](https://github.com/cline/cline):

1.  Open the Cline extension in VS Code.
2.  Click the **MCP Servers** button (server icon).
3.  Click **Configure MCP Servers**. This opens the JSON configuration file.
4.  Add `mcpany` to the `mcpServers` object:

```json
{
  "mcpServers": {
    "mcpany": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v",
        "/absolute/path/to/your/config.yaml:/etc/mcpany/config.yaml",
        "ghcr.io/mcpany/server:latest",
        "run",
        "--config-path",
        "/etc/mcpany/config.yaml",
        "--stdio"
      ]
    }
  }
}
```
