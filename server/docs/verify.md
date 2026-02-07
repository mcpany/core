# System Verification & Diagnostics

This guide details the tools and commands available in MCP Any to verify your configuration, system health, and security integrity.

## 1. Configuration Verification

Before starting the server, it is critical to ensure your configuration files are syntactically correct and logically valid.

### `mcpany config check` (Schema Validation)

This command performs a fast validation of your configuration file against the internal JSON Schema. It checks for:
- Correct YAML/JSON syntax.
- Required fields.
- Correct data types.

```bash
mcpany config check ./config.yaml
```

### `mcpany config validate` (Deep Validation)

This command performs a comprehensive validation of the configuration. In addition to schema checks, it verifies:
- Logical consistency (e.g., referenced profiles exist).
- Environment variable substitution.
- File path existence (for local tools).

```bash
# Basic validation
mcpany config validate --config-path ./config.yaml

# Validate and also attempt to connect to upstream services
mcpany config validate --config-path ./config.yaml --check-connection
```

### `mcpany lint` (Best Practices)

The linter analyzes your configuration for potential issues that aren't strictly errors but might be suboptimal or insecure. It checks for:
- Hardcoded secrets (recommends using environment variables or secret stores).
- Deprecated fields.
- Missing descriptions for tools.

```bash
mcpany lint --config-path ./config.yaml
```

## 2. Connectivity Diagnostics

### `mcpany doctor`

The `doctor` command is your primary tool for diagnosing connectivity and environment issues. It interactively checks:
- **Environment Variables**: Ensures all variables referenced in the config are set. It provides "Actionable Errors" with suggestions for fixes (e.g., typo correction).
- **Network Connectivity**: Attempts to connect to all configured upstream services (HTTP, gRPC, WebSocket).
- **Dependencies**: Verifies that required local executables (e.g., `python3`, `node`) are installed and accessible for command-line tools.

```bash
mcpany doctor --config-path ./config.yaml
```

If the doctor finds issues, it will report them with specific remediation steps.

## 3. Runtime Health

### `mcpany health`

Once the server is running, you can check its health status using the `health` command or by querying the health endpoint directly.

```bash
# Check the health of a running server (default localhost:50050)
mcpany health

# Specify a different address or timeout
mcpany health --mcp-listen-address localhost:8080 --timeout 2s
```

This command queries the server's internal health check logic, which aggregates the status of all subsystems.

## 4. Security Verification (Tool Integrity)

MCP Any includes **Tool Poisoning Mitigation** to prevent "Rug Pull" attacks where an upstream service changes its tool definitions maliciously.

### How it works

When a tool is registered, the server can enforce an integrity check using a SHA256 hash of the tool's definition.
- **Verification**: If a tool definition includes an `integrity` hash, the server calculates the hash of the loaded tool and compares it.
- **Enforcement**: If the hashes do not match, the tool is rejected and not exposed to the agent.

For more details on configuring security features, see [Security Features](features/security.md).
