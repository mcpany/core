# File System Provider

MCP Any allows you to expose local file systems as MCP resources. This enables AI agents to read, write, and list files within allowed directories.

## Configuration

To configure a file system service, use the `filesystem` type in your `services` configuration.

```yaml
services:
  - id: "local-files"
    name: "Local Files"
    type: "filesystem"
    config:
      path: "/var/data"  # The root directory to expose
      read_only: false   # Set to true to prevent write operations
```

## Features

- **List Directory**: Agents can list files and subdirectories.
- **Read File**: Read the content of files (text or binary).
- **Write File**: Create or overwrite files (if `read_only` is false).
- **Delete File**: Remove files (if `read_only` is false).
- **Sandboxing**: Access is restricted to the configured `path`. Agents cannot access files outside this directory (e.g., via `..`).

## Security

- **Path Traversal Prevention**: The server validates all paths to ensure they stay within the root directory.
- **Read-Only Mode**: Recommended for most use cases where agents only need to consume data.
