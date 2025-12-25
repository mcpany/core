# Filesystem Provider

The Filesystem provider allows MCP Any to expose a local directory as an MCP server. This enables LLMs to list directories, read files, write files, and get file information within a controlled scope.

## Features

*   **Sandboxed Access**: Restrict access to specific root directories.
*   **Read/Write Control**: Configure read-only or read-write access.
*   **File Operations**:
    *   `list_directory`: List files and subdirectories.
    *   `read_file`: Read the content of a file.
    *   `write_file`: Write content to a file.
    *   `get_file_info`: Get metadata about a file or directory.
*   **Virtual File Systems**: Supports `os` (local operating system) and `tmpfs` (in-memory). Future support planned for cloud storage.

## Configuration

To use the filesystem provider, add a `filesystem` entry to your `upstream_services` configuration.

```yaml
upstream_services:
  - name: "local-files"
    type: "filesystem"
    filesystem:
      filesystem_type: "os" # or "tmpfs"
      read_only: false
      root_paths:
        "/data": "/var/lib/my-app/data"
        "/logs": "/var/log/my-app"
```

### Parameters

*   `filesystem_type`:
    *   `os`: Maps to the local operating system's filesystem.
    *   `tmpfs`: Creates a temporary in-memory filesystem (useful for testing or ephemeral scratchpads).
*   `read_only`: If `true`, write operations will be disabled.
*   `root_paths`: A map of virtual paths to real paths.
    *   Key: The virtual path exposed to the LLM (e.g., `/data`).
    *   Value: The actual path on the server (e.g., `/mnt/data`).
    *   This provides a chroot-like environment where the LLM can only access files under the specified real paths.

## Security

*   **Path Traversal Protection**: The provider resolves all paths and ensures they stay within the configured `root_paths`. Symlinks are resolved to check the final destination.
*   **Access Control**: Use `read_only` to prevent modification of sensitive files.
