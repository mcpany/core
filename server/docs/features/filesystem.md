# Filesystem Provider

The Filesystem Provider enables MCP Any to safely expose local directories as tools to LLM agents. This allows agents to read, write, and list files within specific, sandboxed directories.

## Configuration

To enable the filesystem provider, add a `filesystem_service` entry to your `mcpany` configuration.

```yaml
services:
  - name: my-workspace
    filesystem_service:
      root_paths:
        "/workspace": "/home/user/projects/myproject"
        "/logs": "/var/log/myapp"
      read_only: false
```

### Fields

*   `root_paths` (map<string, string>): A mapping of **virtual paths** (exposed to the LLM) to **real local paths**.
    *   The key is the virtual path (e.g., `/workspace`).
    *   The value is the absolute path on the local machine (e.g., `/home/user/projects/myproject`).
    *   **Security:** Access is strictly limited to these directories and their subdirectories. Symbolic links pointing outside these roots are blocked.
*   `read_only` (bool): If set to `true`, the agent can only read and list files. Write operations (`write_file`) will fail.

## Tools

The following tools are automatically registered:

| Tool Name | Description | Inputs | Outputs |
| :--- | :--- | :--- | :--- |
| `list_directory` | Lists files and subdirectories in a given path. | `path` (string) | List of entries (name, is_dir, size). |
| `read_file` | Reads the content of a file. | `path` (string) | `content` (string). |
| `write_file` | Writes content to a file. Overwrites if exists. | `path` (string), `content` (string) | `success` (boolean). |
| `get_file_info` | Gets metadata about a file or directory. | `path` (string) | Metadata (name, is_dir, size, mod_time). |
| `list_allowed_directories` | Lists the configured virtual root paths. | None | List of allowed roots. |

## Security Considerations

1.  **Sandboxing:** The provider enforces strict path validation. Attempts to access paths like `../../etc/passwd` or follow symlinks pointing outside the allowed roots will result in an "access denied" error.
2.  **Read-Only Mode:** Use `read_only: true` for scenarios where the agent should analyze code or logs but not modify them.
3.  **Authentication:** Like all MCP Any services, access to these tools can be controlled via RBAC and authentication policies defined in the `authentication` section of your config.

## Example

**Scenario:** You want an agent to help you refactor code in your current project.

1.  Run `mcpany` with the following config:

    ```yaml
    services:
      - name: code-refactor-agent
        filesystem_service:
          root_paths:
            "/src": "."
          read_only: false
    ```

2.  The agent can now:
    *   List files: `list_directory(path="/src")`
    *   Read code: `read_file(path="/src/main.go")`
    *   Apply fixes: `write_file(path="/src/main.go", content="...")`
