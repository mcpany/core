# Server Configuration Guide

**Goal**: Set up a robust, secure MCP Any Server.

## Pain Point: "Where do I start?"

The server is driven by a YAML configuration file. By default, it looks for `config.yaml` in the current directory.

### Quick Start
1. Create `config.yaml`:
   ```yaml
   global_settings:
     bind_address: ":50050"
     log_level: "info"

   upstream_services:
     - name: "my-local-tool"
       type: "stdio"
       command: "python3"
       args: ["/path/to/script.py"]
   ```
2. Run the server:
   ```bash
   ./mcp-server run
   ```

## Pain Point: "How do I keep secrets safe?"

Never hardcode API keys in `config.yaml`. Use environment variable substitution.

### Usage
```yaml
upstream_services:
  - name: "github-api"
    type: "http"
    endpoint: "https://api.github.com"
    environment:
      # Will look for GITHUB_TOKEN in the server's environment
      GITHUB_TOKEN: "${GITHUB_TOKEN}"
```

## Pain Point: "My Config is huge!"

Split your configuration using `imports` or directory scanning (if supported) or keep it modular by functionality.

### Best Practices
- **Global Settings** at the top.
- **Group Services** by team or domain (comments help).
- **Use Validation**: Run `mcpctl config validate config.yaml` before restarting.

## Configuration Validation

The server includes advanced validation features to help catch errors early:

- **Actionable Configuration Errors**: The server provides specific "Fix" suggestions for common issues like missing environment variables or invalid paths.
- **Environment Variable Fuzzy Matching**: If you misspell an environment variable (e.g., `OPENAI_API_KEY` vs `OPENAI_APIKEY`), the server will suggest the correct variable name if it finds a close match in the environment.
- **Whitespace URL Validation**: The server automatically detects and warns about hidden leading/trailing whitespace in URLs, which is a common copy-paste error.
