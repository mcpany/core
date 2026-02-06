# Verification Guide

This guide explains how to verify that your MCP Any server is installed and running correctly.

## 1. Using the Doctor Command

The `doctor` command is the primary tool for verifying your installation and configuration. It checks connectivity to upstream services, validates environment variables, and ensures the system is healthy.

### Running Doctor

If you are running the server locally:

```bash
# If installed via go install or built from source
mcpany doctor

# OR running directly from source
go run ./cmd/server doctor
```

If you are using Docker:

```bash
docker run --rm -v $(pwd)/config.yaml:/etc/mcpany/config.yaml ghcr.io/mcpany/server:latest doctor --config-path /etc/mcpany/config.yaml
```

### Expected Output

The command will output a checklist of services and their status. Look for all green checks.

```text
✔ Configuration loaded
✔ Environment variables validated
✔ Upstream services reachable
✔ Local tools discovered
```

## 2. Health API Endpoint

You can check the server's liveness by querying the `/health` endpoint.

```bash
curl http://localhost:50050/health
```

**Expected Response:** `OK` (HTTP 200)

## 3. Dashboard Verification

If you have the UI enabled, navigate to the Dashboard to verify the system status visually.

1.  Open your browser to `http://localhost:3000` (or your configured UI port).
2.  Check the **System Health** widget on the dashboard. It should show a historical timeline of "Healthy" (Green) status.
3.  Navigate to the **Services** page. All configured services should show a green "Active" indicator.

## 4. Simple Tool Test

To verify that the MCP protocol is working, you can try listing the available tools.

If you have `curl` or a similar tool, you can send a JSON-RPC request to the HTTP endpoint (if enabled):

```bash
curl -X POST http://localhost:50050/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 1
  }'
```

**Expected Response:** A JSON object containing a list of tools.

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [ ... ]
  }
}
```
