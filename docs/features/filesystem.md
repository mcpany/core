# File System Provider

MCP Any allows you to expose local file systems as MCP resources. This enables AI agents to read, write, and list files within allowed directories.

## Configuration

To configure a file system service, use the `filesystem` type in your `services` configuration.

```yaml
upstream_services:
  - name: "Local Files"
    filesystem_service:
      root_paths:
        "/data": "/var/data" # Map virtual /data to local /var/data
      os: {}
      read_only: false   # Set to true to prevent write operations
```

### S3 Configuration

You can also configure an S3 bucket as a filesystem.

```yaml
upstream_services:
  - name: "My Bucket"
    filesystem_service:
      s3:
        bucket: "my-bucket"
        region: "us-east-1"
        # Optional: Credentials (if not using env vars or instance profile)
        # access_key_id: "..."
        # secret_access_key: "..."
      read_only: true
```

### GCS Configuration

You can configure a Google Cloud Storage bucket as a filesystem.

```yaml
upstream_services:
  - name: "My GCS Bucket"
    filesystem_service:
      gcs:
        bucket: "my-gcs-bucket"
      read_only: true
```

## Features

- **List Directory**: Agents can list files and subdirectories.
- **Read File**: Read the content of files (text or binary).
- **Write File**: Create or overwrite files (if `read_only` is false).
- **Delete File**: Remove files (if `read_only` is false).
- **Search Files**: Search for a text pattern (regex) in files within a directory.
- **Get File Info**: Get metadata (size, mod time, is_dir) about a file or directory.
- **Sandboxing**: Access is restricted to the configured `path`. Agents cannot access files outside this directory (e.g., via `..`).

## Security

- **Path Traversal Prevention**: The server validates all paths to ensure they stay within the root directory.
- **Read-Only Mode**: Recommended for most use cases where agents only need to consume data.
