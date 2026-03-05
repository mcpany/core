# Market Sync: 2026-03-05

## Ecosystem Updates

### OpenClaw & Dynamic Permissions
- **JIT Permission Shifts**: OpenClaw is moving towards a model where agent permissions are not just static "scopes" defined at startup, but dynamic "leases" granted just-in-time. This reduces the risk of long-lived, overly-broad access tokens.
- **Headless Authorization**: Introduction of "Authorization Servers" that agents can query to receive temporary cryptographic attestations for specific tool calls.

### Claude Code & Gemini CLI
- **Interactive Permission Flows**: Claude Code's latest update includes a structured way for the agent to ask the user: "I need to write to `package.json` to update a dependency. Do you allow this once / always / never?".
- **Contextual Sandboxing**: Increased use of ephemeral containers that are spun up and torn down for a single "task session," further isolating tool execution.

## Security & Vulnerabilities

### The "Leaky Tool" Problem
- **PII Leakage in Tool Outputs**: Several community-contributed MCP servers were found to be leaking AWS metadata and internal environment variables in their `stderr` or `logs` outputs, which LLMs then ingested and occasionally echoed back to users.
- **Redaction Middleware Necessity**: Growing demand for a standardized way to "scrub" tool outputs (stdout/stderr) for secrets and PII before they are returned to the agent.

## Autonomous Agent Pain Points
- **Consent Fatigue**: Users are reporting "consent fatigue" from too many HITL prompts. There is a need for "Policy-Driven JIT" where certain safe escalations are pre-approved by the organization.
- **Blast Radius Management**: Agents still lack a way to "self-restrict" their own capabilities when performing high-risk tasks, leading to accidental destructive actions.
