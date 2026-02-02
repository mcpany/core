# Server Configuration Guide

**Goal**: Set up a robust, secure MCP Any Server.

## Pain Point: "Where do I start?"

The server is driven by a YAML configuration file. By default, it looks for `config.yaml` in the current directory.

### Quick Start
1. Create `config.yaml`:
   ```yaml
   global_settings:
     mcp_listen_address: ":50050"
     log_level: "info"

   upstream_services:
     - name: "my-local-tool"
       mcp_service:
         stdio_connection:
           command: "python3"
           args: ["/path/to/script.py"]
   ```
2. Run the server:
   ```bash
   ./server run --config-path config.yaml
   ```

## Pain Point: "How do I keep secrets safe?"

Never hardcode API keys in `config.yaml`. Use environment variable substitution.

### Usage
```yaml
upstream_services:
  - name: "github-api"
    http_service:
      address: "https://api.github.com"
    upstream_auth:
      api_key:
        param_name: "Authorization"
        in: HEADER
        value:
          plain_text: "Bearer ${GITHUB_TOKEN}" # Will look for GITHUB_TOKEN in the server's environment
```

## Pain Point: "My Config is huge!"

Split your configuration using `imports` or directory scanning (if supported) or keep it modular by functionality.

### Best Practices
- **Global Settings** at the top.
- **Group Services** by team or domain (comments help).
- **Use Validation**: Run `mcpctl config validate config.yaml` before restarting.
