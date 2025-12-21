# Filesystem Provider

The Filesystem provider allows you to expose a local directory as a set of MCP tools. This is useful for giving your AI agent access to read and write files in a specific workspace.

## Configuration

To use the filesystem provider, add a `filesystem_service` entry to your configuration.

```yaml
upstream_services:
  - name: "my-project-files"
    filesystem_service:
      root_path: "/path/to/project"
      read_only: false
```

### Configuration Options

| Field | Type | Description | Required | Default |
| :--- | :--- | :--- | :--- | :--- |
| `root_path` | string | The absolute or relative path to the directory you want to expose. | Yes | - |
| `read_only` | bool | If true, the `write_file` tool will not be registered, preventing any modifications. | No | `false` |

## Tools

The following tools are automatically registered:

### `list_files`
Lists files and directories in the specified path (relative to `root_path`).
- **Inputs**:
  - `path` (string, optional): The directory to list. Defaults to `.`.

### `read_file`
Reads the content of a file.
- **Inputs**:
  - `path` (string): The file to read (relative to `root_path`).

### `write_file`
Writes content to a file. (Only available if `read_only` is false).
- **Inputs**:
  - `path` (string): The file to write (relative to `root_path`).
  - `content` (string): The content to write.

### `get_file_info`
Gets metadata about a file or directory.
- **Inputs**:
  - `path` (string): The path to check.

## Security

- **Path Traversal Protection**: The provider strictly enforces that all file operations occur within the `root_path`. Attempts to access files outside (e.g., using `../`) will result in an error.
- **Permissions**: Files are written with `0600` permissions (readable/writable only by the owner).
