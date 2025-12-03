
# Filesystem Service Example

This example demonstrates how to configure and use the filesystem service in MCP Any.

## Configuration

The following configuration exposes the local filesystem as a set of tools. The `basePath` is set to `/tmp`, so all file operations will be relative to that directory.

```yaml
# examples/filesystem/config.yaml
upstreamServices:
  - name: "my-filesystem-service"
    filesystemService:
      basePath: "/tmp"
```

## Usage

To run the server with this configuration, use the following command:

```bash
make run ARGS="--config-paths ./examples/filesystem/config.yaml"
```

### List Files

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-filesystem-service/-/listFiles", "arguments": {"path": "."}}, "id": 1}' \
  http://localhost:50050
```

### Write File

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-filesystem-service/-/writeFile", "arguments": {"path": "test.txt", "content": "Hello, world!"}}, "id": 2}' \
  http://localhost:50050
```

### Read File

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-filesystem-service/-/readFile", "arguments": {"path": "test.txt"}}, "id": 3}' \
  http://localhost:50050
```

### Delete File

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-filesystem-service/-/deleteFile", "arguments": {"path": "test.txt"}}, "id": 4}' \
  http://localhost:50050
```
