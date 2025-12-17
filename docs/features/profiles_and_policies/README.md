# Profiles and Call Policies

MCP Any provides powerful mechanisms to control which tools are exposed to the AI agent and how they can be used. This is crucial for managing complexity, ensuring security, and tailoring the agent's environment.

## Strategies

There are two primary strategies for controlling access:

1.  **Call Policy**: Best for restricting specific APIs within a service, especially when auto-discovery logic (like OpenAPI or gRPC reflection) exposes too many tools.
2.  **Profiles**: Best for creating different "views" of the system for different environments (Dev vs Prod) or different types of agents.

---

## 1. Call Policy (Restricting APIs)

When you enable `auto_discover_tool: true` for an OpenAPI or gRPC service, it might expose hundreds of endpoints. You may want to restrict the AI agent to only a safe subset.

### Configuration

Call Policies are defined under `upstream_services`. You can set a `default_action` (ALLOW or DENY) and then provide specific exceptions.

#### Example: Default Deny (Whitelist)

This is the most secure approach. Only explicitly allowed tools are exposed.

```yaml
upstream_services:
  - name: "huge-api"
    openapi_service:
      spec_url: "https://api.example.com/openapi.json"
    auto_discover_tool: true
    call_policies:
      # Block everything by default
      - default_action: DENY
        rules:
          # Allow searching users
          - action: ALLOW
            name_regex: "^searchUsers$"
          # Allow getting user details
          - action: ALLOW
            name_regex: "^getUser$"
```

#### Example: Default Allow (Blacklist)

Use this if you generally trust the service but want to block specific dangerous actions.

```yaml
upstream_services:
  - name: "backend-service"
    grpc_service:
      address: "localhost:50051"
      use_reflection: true
    call_policies:
      - default_action: ALLOW
        rules:
          # Block any delete operations
          - action: DENY
            name_regex: "^Delete.*"
          # Block admin endpoints
          - action: DENY
            name_regex: ".*AdminService.*"
```

---

## 2. Profiles (Agent-Specific Views)

Profiles allow you to group services or resources and selectively enable them at runtime. This is perfect for presenting different toolsets to different agents.

### Configuration

You tag each upstream service with one or more profiles. Note that you should set both `id` and `name` for the profile to ensure consistent behavior across the system.

```yaml
global_settings:
  profiles:
    - "dev"
    - "prod"

upstream_services:
  - name: "debug-tool"
    profiles:
      - id: "dev"
        name: "dev"
    command_line_service:
      command: "echo"
      args: ["debug"]

  - name: "payment-processor"
    profiles:
      - id: "prod"
        name: "prod"
      - id: "staging"
        name: "staging"
    http_service:
      address: "https://api.stripe.com"
```

### Usage

When starting the MCP Any server, you specify which profiles are active. This can be done via `global_settings` in the config file (recommended) or via command line flags.

- **Development Agent**:

  ```bash
  mcp-server run --config config.yaml --profiles=dev
  ```

  _Sees: `debug-tool`_

- **Production Agent**:
  ```bash
  mcp-server run --config config.yaml --profiles=prod
  ```
  _Sees: `payment-processor`_

### Resources and Prompts

Profiles also filter **Resources** and **Prompts** associated with the service. If a service is visible to a profile, all its resources and prompts are also visible to that profile (unless further restricted by Call Policy).

You can combine both strategies! Use **Profiles** to select high-level service sets for an environment, and **Call Policies** to fine-tune exactly which methods on those services are safe to call.

---

## Examples & Verification

We provide End-to-End tests demonstrating these patterns.

### Restricted API (Call Policy)

Demonstrates denying all tools by default and allowing only a specific one.
[View Example](./examples/restricted_api)

### Agent Profiles

Demonstrates switching between "planning" and "executor" profiles to change available tools.
[View Example](./examples/agent_profiles)
