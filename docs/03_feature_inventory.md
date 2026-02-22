# Feature Inventory

## Status: Active & Proposed

### Core Infrastructure
- **Universal Adapter (HTTP/gRPC/CLI):** [Status: Implemented]
- **Dynamic Configuration Hot-Reload:** [Status: Implemented]
- **Service Registry:** [Status: Implemented]

### Security & Governance
- **API Key & Bearer Auth:** [Status: Implemented]
- **DLP Redaction:** [Status: Implemented]
- **[New] Zero Trust Policy Firewall:** [Status: Proposed] - Granular runtime tool call validation using Rego/CEL.

### Advanced Agent Capabilities
- **Context Optimizer:** [Status: Implemented] - Token usage reduction.
- **[New] Recursive Context Protocol:** [Status: Proposed] - Standardized headers for parent-to-subagent context inheritance.
- **[New] Shared KV Blackboard:** [Status: Proposed] - SQLite-backed shared memory for agent swarms.

### Observability
- **Audit Logging:** [Status: Implemented]
- **Health Dashboard:** [Status: Implemented]
- **[New] Tool Execution Timeline:** [Status: Proposed] - Visual waterfall of tool call stages.

## Priority Shifts: 2026-02-22
- **Escalating "Recursive Context Protocol" to P1:** Critical for OpenClaw subagent support.
- **Escalating "Zero Trust Policy Firewall" to P0:** Essential to mitigate rising prompt injection vulnerabilities in autonomous swarms.
