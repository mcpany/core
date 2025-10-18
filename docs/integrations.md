# 🔌 Integrating with AI Assistants

This guide explains how to connect `mcpxy` to various AI assistant clients, enabling them to leverage the tools exposed by your `mcpxy` server. By integrating `mcpxy`, you can give your AI assistant access to gRPC services, REST APIs, local command-line tools, and more.

There are three primary ways to run `mcpxy` for integration:

1.  **As a Local Binary**: Run the `mcpxy` server directly on your machine.
2.  **With Docker Compose**: Run `mcpxy` as a containerized service.
3.  **With Helm**: Deploy `mcpxy` to a Kubernetes cluster.

---

## 1. Local Binary Integration

This is the most direct method for running `mcpxy` on your local machine.

### Prerequisites

Before you begin, you need to have the `mcpxy` server binary built and available on your system.

1.  **Clone the `mcpxy` repository:**

    ```bash
    git clone https://github.com/mcpxy/core.git
    cd core
    ```

2.  **Build the application:**
    This command compiles the server and places the executable binary in the `build/bin` directory.
    ```bash
    make build
    ```
    The server binary will be located at `./build/bin/server`. You will need to use the **absolute path** to this binary when configuring your AI assistant.

### AI Client Setup

Most AI clients require you to specify the command to start the `mcpxy` server.

#### General JSON Configuration

This generic JSON configuration can be adapted for most clients that support it (e.g., VS Code, JetBrains, Cursor).

```json
{
  "mcpServers": {
    "mcpxy": {
      "command": "./build/bin/server",
      "args": ["--config-paths", "/path/to/your/mcpxy-config.yaml"]
    }
  }
}
```

- **Note**: The server binary is now located at `./build/bin/server`.

#### Gemini CLI

To register `mcpxy` as an extension to the Gemini CLI:

```bash
# Add for the current project
gemini mcp add mcpxy "./build/bin/server"

# Add with command-line arguments (like a config file)
# Note the use of '--' to separate the command from its arguments.
gemini mcp add mcpxy -- "./build/bin/server" --config-paths "/path/to/your/mcpxy-config.yaml"
```

#### Claude CLI

```bash
claude mcp add mcpxy "./build/bin/server"
```

#### Copilot CLI

Use the interactive prompt (`/mcp add`) and provide the absolute path to the `server` binary and any optional arguments.

---

## 2. Docker Compose Integration

You can run `mcpxy` and its upstream services in a containerized environment using Docker Compose. This is ideal for creating reproducible setups.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) are installed.

### Running with Docker Compose

The repository includes a `docker-compose.yml` file that starts `mcpxy` and a sample upstream service.

1.  **Start the services:**

    ```bash
    docker compose up --build
    ```

    The `mcpxy` server will be available on `localhost:50050`.

2.  **Configure your AI Assistant:**
    Since `mcpxy` is running as a networked service (not a local process), you need to configure your AI assistant to connect to its TCP port.

    For clients like Gemini, you can add it as an HTTP extension:

    ```bash
    # The 'mcp' subcommand is for local processes. For networked services, use 'http'.
    gemini http add mcpxy-docker http://localhost:50050
    ```

    For clients with a UI (like VS Code or JetBrains), look for an option to add an "HTTP" or "Remote" MCP server and point it to `http://localhost:50050`.

---

## 3. Helm Integration

For production or staging environments, you can deploy `mcpxy` to a Kubernetes cluster using the provided Helm chart.

### Prerequisites

- A running Kubernetes cluster.
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured to connect to your cluster.
- [Helm](https://helm.sh/docs/intro/install/) installed.

### Deployment

1.  **Install the Helm Chart:**
    Navigate to the `helm` directory in this repository and install the chart.

    ```bash
    cd helm/
    helm install mcpxy . -f values.yaml
    ```

    This will deploy `mcpxy` into your Kubernetes cluster.

2.  **Expose the Service:**
    To access `mcpxy` from your local machine, you can use `kubectl port-forward`.

    ```bash
    # Forward a local port (e.g., 50050) to the mcpxy service in the cluster
    kubectl port-forward svc/mcpxy 50050:50050
    ```

3.  **Configure your AI Assistant:**
    With the port forward running, `mcpxy` is now accessible at `localhost:50050`. You can configure your AI assistant the same way as in the Docker Compose example.

    ```bash
    gemini http add mcpxy-k8s http://localhost:50050
    ```

By following these instructions, you can connect `mcpxy` to your favorite AI coding assistant, regardless of how you choose to run it.
