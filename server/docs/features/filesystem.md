# Filesystem Provider

The Filesystem Provider allows the MCP Any server to expose a local or remote filesystem as a set of tools. This enables LLMs to read, write, list, and search files securely within configured root directories.

## Configuration

The filesystem provider is configured as an `upstream_service` with the `filesystem` block.

### Fields

| Field        | Type                  | Description                                                                 |
| ------------ | --------------------- | --------------------------------------------------------------------------- |
| `root_paths` | `map<string, string>` | A map where keys are virtual paths (exposed to LLM) and values are real paths. |
| `read_only`  | `bool`                | If true, disables write and delete operations.                              |

### Supported Backends

The `filesystem` block supports multiple backend types. You must specify exactly one.

#### 1. Local Filesystem (OS)

```yaml
upstream_services:
  - name: "local-fs"
    filesystem:
      root_paths:
        "/workspace": "/home/user/projects/myproject"
      read_only: false
      os: {}
```

#### 2. S3 (Amazon Simple Storage Service)

```yaml
upstream_services:
  - name: "s3-fs"
    filesystem:
      root_paths:
        "/data": "/" # Map root of bucket to /data
      s3:
        bucket: "my-bucket"
        region: "us-east-1"
        access_key_id: "${AWS_ACCESS_KEY_ID}"
        secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
```

#### 3. GCS (Google Cloud Storage)

```yaml
upstream_services:
  - name: "gcs-fs"
    filesystem:
      gcs:
        bucket: "my-gcs-bucket"
```

#### 4. ZIP Archive

```yaml
upstream_services:
  - name: "archive-fs"
    filesystem:
      zip:
        file_path: "/path/to/archive.zip"
      read_only: true
```

#### 5. SFTP

```yaml
upstream_services:
  - name: "sftp-fs"
    filesystem:
      sftp:
        address: "sftp.example.com:22"
        username: "user"
        password: "password"
        # Or use key_path
        # key_path: "/path/to/private/key"
```

## Tools Exposed

When a filesystem service is configured, it automatically registers the following tools:

- `list_directory`: List files and directories in a given path.
- `read_file`: Read the content of a file.
- `write_file`: Write content to a file (disabled if `read_only` is true).
- `delete_file`: Delete a file or empty directory (disabled if `read_only` is true).
- `search_files`: Search for a text pattern (regex) in files within a directory.
- `get_file_info`: Get metadata (size, mod time) about a file or directory.

## Security Considerations

- **Sandboxing**: Access is strictly limited to the directories specified in `root_paths`. Attempts to access files outside these roots (e.g., via `..`) are blocked.
- **Read-Only Mode**: Use `read_only: true` for sensitive data that should not be modified by the LLM.
- **File Size Limits**: `read_file` enforces a 10MB limit to prevent memory exhaustion.
- **Hidden Files**: `search_files` skips hidden directories (starting with `.`) by default.

## Example Usage

**Configuration:**

```yaml
upstream_services:
  - name: "project-files"
    filesystem:
      root_paths:
        "/src": "./src"
      read_only: false
      os: {}
```

**LLM Interaction:**

User: "List files in the source directory."

Model calls `project-files` -> `list_directory` with arguments `{"path": "/src"}`.

**Response:**

```json
{
  "entries": [
    {"name": "main.go", "is_dir": false, "size": 1024},
    {"name": "utils", "is_dir": true, "size": 4096}
  ]
}
```
